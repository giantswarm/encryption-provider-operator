package encryption

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/go-logr/logr"
	"golang.org/x/crypto/sha3"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	configv1 "github.com/giantswarm/encryption-provider-operator/pkg/config"
	"github.com/giantswarm/encryption-provider-operator/pkg/key"
	"github.com/giantswarm/encryption-provider-operator/pkg/project"
)

const (
	EncryptionProviderConfig = "encryption"
	KeyNamePrefix            = "key"

	// Poly1305KeyLength represents the 32 bytes length for Poly1305
	// padding encryption key.
	Poly1305KeyLength = 32

	AnnotationRotationInProgress = "encryption.giantswarm.io/rotation-in-progress"
	AnnotationForceRotation      = "encryption.giantswarm.io/force-rotation"
	AnnotationEnableRotation     = "encryption.giantswarm.io/enable-rotation"
	AnnotationRewriteTimestamp   = "encryption.giantswarm.io/rewrited-at"
	AnnotationLastRotation       = "encryption.giantswarm.io/last-rotation"

	EncryptionProviderConfigShake256SecretName      = "encryption-provider-config-shake256"
	EncryptionProviderConfigShake256SecretNamespace = "kube-system"

	MasterNodeLabel = "node-role.kubernetes.io/master"
)

type Config struct {
	Cluster                  *capi.Cluster
	DefaultKeyRotationPeriod time.Duration
	RegistryDomain           string

	CtrlClient ctrlclient.Client
	Logger     logr.Logger
}

type Service struct {
	cluster                  *capi.Cluster
	defaultKeyRotationPeriod time.Duration
	registryDomain           string

	ctrlClient ctrlclient.Client
	logger     logr.Logger
}

func New(c Config) (*Service, error) {
	if c.Cluster == nil {
		return nil, errors.New("cluster cannot be nil")
	}
	if c.CtrlClient == nil {
		return nil, errors.New("ctrlClient cannot be nil")
	}
	if c.RegistryDomain == "nil" {
		return nil, errors.New("RegistryDomain cannot be empty")
	}
	if c.Logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	s := &Service{
		cluster:                  c.Cluster,
		registryDomain:           c.RegistryDomain,
		defaultKeyRotationPeriod: c.DefaultKeyRotationPeriod,
		ctrlClient:               c.CtrlClient,
		logger:                   c.Logger,
	}

	return s, nil
}

func (s *Service) Reconcile() error {
	ctx := context.TODO()
	clusterName := s.cluster.ClusterName
	if clusterName == "" {
		clusterName = s.cluster.Name
	}

	var encryptionProviderSecret v1.Secret

	err := s.ctrlClient.Get(ctx, ctrlclient.ObjectKey{
		Name:      key.SecretName(clusterName),
		Namespace: s.cluster.Namespace,
	}, &encryptionProviderSecret)

	if apierrors.IsNotFound(err) {
		// create new encryption secret
		err := s.createNewEncryptionProviderSecret(ctx, clusterName)
		if err != nil {
			s.logger.Error(err, "failed to get encryption provider config secret for cluster")
			return err
		}
	} else if err != nil {
		s.logger.Error(err, "failed to get encryption provider config secret for cluster")
		return err
	} else {
		// config already exists, check for key rotation
		err = s.keyRotation(ctx, encryptionProviderSecret, clusterName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Delete() error {
	ctx := context.TODO()
	clusterName := s.cluster.ClusterName
	if clusterName == "" {
		clusterName = s.cluster.Name
	}
	encryptionProviderSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.SecretName(clusterName),
			Namespace: s.cluster.Namespace,
		}}

	err := s.ctrlClient.Delete(ctx, &encryptionProviderSecret)
	if apierrors.IsNotFound(err) {
		// secret is already deleted, fall thru
		return nil
	} else if err != nil {
		s.logger.Error(err, "failed to delete encryption provider config secret for cluster")
		return err
	}

	err = key.CleanWCK8sKubeconfig(clusterName)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to delete local kubeconfig file for cluster %s", clusterName))
		return err
	}

	return nil
}

func (s *Service) createNewEncryptionProviderSecret(ctx context.Context, clusterName string) error {
	// check if there is old encryption config that we can use for migration
	oldEncryptionSecretName := fmt.Sprintf("%s-encryption", clusterName)

	var oldEncryptionSecret v1.Secret
	err := s.ctrlClient.Get(ctx, ctrlclient.ObjectKey{
		Name:      oldEncryptionSecretName,
		Namespace: s.cluster.Namespace,
	},
		&oldEncryptionSecret)

	var encryptionConfig configv1.EncryptionConfiguration

	if apierrors.IsNotFound(err) {
		// no old key found, lets generate a new one
		newKey, err := newRandomKey(Poly1305KeyLength)
		if err != nil {
			s.logger.Error(err, "failed to generate new random key for encryption")
			return err
		}
		s.logger.Info("generated a new encryption key for Poly1305 encryption provider")

		providerConfig := configv1.ProviderConfiguration{
			Secretbox: &configv1.SecretboxConfiguration{
				Keys: []configv1.Key{
					{
						Name:   "key1",
						Secret: newKey,
					},
				},
			},
		}
		encryptionConfig = initNewEncryptionConfigStruct(providerConfig)
	} else if err != nil {
		s.logger.Error(err, "failed to get old encryption provider key secret")
		return err
	} else {
		// there is an old encryption key so lets reuse it to avoid breaking cluster
		if k, ok := oldEncryptionSecret.Data["encryption"]; !ok {
			s.logger.Error(err, "failed to get encryption key from secret")
			return err
		} else {
			// the old encryption key are using old less secure type
			// https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#providers
			providerConfig := configv1.ProviderConfiguration{
				AESCBC: &configv1.AESConfiguration{
					Keys: []configv1.Key{
						{
							Name:   "key1",
							Secret: string(k),
						},
					},
				},
			}
			encryptionConfig = initNewEncryptionConfigStruct(providerConfig)
			s.logger.Info("fetched and migrated AESCBC encryption key from legacy product")
		}
	}

	secretData, err := yaml.Marshal(&encryptionConfig)
	if err != nil {
		return err
	}

	encryptionProviderSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.SecretName(clusterName),
			Namespace: s.cluster.Namespace,
			Labels: map[string]string{
				label.Cluster:        clusterName,
				label.ManagedBy:      project.Name(),
				key.ClusterNameLabel: clusterName,
			},
		},
		Data: map[string][]byte{EncryptionProviderConfig: secretData},
	}

	err = s.ctrlClient.Create(ctx, encryptionProviderSecret)
	if err != nil {
		s.logger.Error(err, "failed to create encryption provider secret")
		return err
	}

	s.logger.Info("created a new encryption provider config secret")

	return nil
}

// keyRotation will handle encryption key rotation in case the configured time period elapsed
// the controller needs to handle several phases of the rotation as it is require roll of the master nodes and rewriting all the secrets
//
func (s *Service) keyRotation(ctx context.Context, encryptionProviderSecret v1.Secret, clusterName string) error {
	// check if key rotation is already in progress
	if _, ok := encryptionProviderSecret.Annotations[AnnotationRotationInProgress]; ok {
		/*
			short description fo what we do here
			- we get the k8s client to workload cluster
			- the workload cluster should run app that run pod on each master node
			and check the encryption config file on host  and save its md5 sum to the secret called EncryptionProviderConfigShake256SecretName
			this secret will then have md5 check sum fo the file for each master
			- we fetch teh secret and compare the values of md5 checksum with expected value
			- this has to match for each master node
			- in case all master nodes has same new config file we can rewrite all secrets in the workload cluster
		*/
		// get workload cluster k8s client
		wcClient, err := key.GetWCK8sClient(ctx, s.ctrlClient, clusterName, s.cluster.Namespace)
		if err != nil {
			return err
		}

		// calculate md5 checksum of the encryption provider config file
		configShakeSum := shake256Sum(encryptionProviderSecret.Data[EncryptionProviderConfig])
		masterNodesUpToDate, err := s.countMasterNodesWithLatestConfig(ctx, wcClient, configShakeSum)
		if err != nil {
			return err
		}

		if masterNodesUpToDate {
			// rewrite all secrets in workload cluster so new keys is used for encryption
			err = rewriteAllSecrets(wcClient, ctx)
			if err != nil {
				s.logger.Error(err, "failed to rewrite all secrets in workload cluster cluster")
				return err
			}
			s.logger.Info("all secrets on the workload cluster has been rewritten with the new encryption key")

			err = removeOldEncryptionKey(&encryptionProviderSecret)
			if err != nil {
				s.logger.Error(err, "failed to remove old encryption key from the configuration secret")
				return err
			}

			encryptionProviderSecret.Annotations[AnnotationLastRotation] = time.Now().Format(time.RFC3339)
			delete(encryptionProviderSecret.Annotations, AnnotationRotationInProgress)
			err = s.ctrlClient.Update(ctx, &encryptionProviderSecret)
			if err != nil {
				s.logger.Error(err, "failed to update encryption provider secret")
				return err
			}

			// TODO
			// delete the app that watches the encryption config

		}
		// key rotation is not in progress
		// check if the rotation should be started
	} else if _, ok := encryptionProviderSecret.Annotations[AnnotationEnableRotation]; ok {
		addNewKeyForRotation := false

		if t, ok := encryptionProviderSecret.Annotations[AnnotationLastRotation]; ok {
			lastRotation, err := time.Parse(time.RFC3339, t)
			if err != nil {
				s.logger.Error(err, "failed to parse time for last rotation")
				return err
			}

			if time.Since(lastRotation) > s.defaultKeyRotationPeriod {
				addNewKeyForRotation = true
			}

			// the annotation regarding last rotation is missing so assume this is new cluster
			// use creation timestamp to calculate elapsed time
		} else if time.Since(encryptionProviderSecret.CreationTimestamp.Time) > s.defaultKeyRotationPeriod {
			addNewKeyForRotation = true
		}

		// check if the key rotation is forced by the annotation
		if _, ok := encryptionProviderSecret.Annotations[AnnotationForceRotation]; ok {
			addNewKeyForRotation = true
		}

		if addNewKeyForRotation {
			// generate new encryption key
			newKey, err := newRandomKey(Poly1305KeyLength)
			if err != nil {
				s.logger.Error(err, "failed to generate new encryption key")
				return err
			}
			err = addNewEncryptionKey(&encryptionProviderSecret, newKey)
			if err != nil {
				s.logger.Error(err, "failed to add new encryption key to the configuration secret")
				return err
			}

			// keys added, set the new phase on the object
			encryptionProviderSecret.Annotations[AnnotationRotationInProgress] = "yes"
			// delete the Force rotation annotation if it exists
			delete(encryptionProviderSecret.Annotations, AnnotationForceRotation)

			// update the object
			err = s.ctrlClient.Update(ctx, &encryptionProviderSecret)
			if err != nil {
				s.logger.Error(err, "failed to update encryption provider secret")
				return err
			}

			// TODO
			// deploy the app that watches the encryption config

		} else {
			s.logger.Info("keys are not %s old, not rotating")
		}

	} else {
		s.logger.Info("key rotation is not enabled for this secret, skipping")
	}

	return nil
}

// newRandomKey generates a new random key with defined length
func newRandomKey(length int) (string, error) {
	randomKey := make([]byte, length)

	_, err := rand.Read(randomKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(randomKey), nil
}

// addNewProviderKey will add a new key to the config in a way that its always on the first position
// only key on the first position is used to write new secrets to the storage
// if the same encryption provider already exists it will add the key to it otherwise it will ad the new provider at the start
func addNewEncryptionKey(secret *v1.Secret, newEncryptionKey string) error {
	var ec configv1.EncryptionConfiguration
	err := yaml.Unmarshal(secret.Data[EncryptionProviderConfig], &ec)
	if err != nil {
		return err
	}

	added := false
	for _, p := range ec.Resources[0].Providers {
		if p.Secretbox != nil {
			i, err := getMaxKeyIndex(p.Secretbox.Keys)
			if err != nil {
				return err
			}
			// secretbox configuration exists add a new key at the start of the array
			p.Secretbox.Keys = append([]configv1.Key{{Secret: newEncryptionKey, Name: keyName(i + 1)}}, p.Secretbox.Keys...)
			added = true
			break
		}
	}

	// provider secret box is not yet present in the config so add the whole configuration
	if !added {
		secretBoxProvider := configv1.ProviderConfiguration{
			Secretbox: &configv1.SecretboxConfiguration{
				Keys: []configv1.Key{
					{
						Name:   keyName(1),
						Secret: newEncryptionKey,
					},
				},
			},
		}
		ec.Resources[0].Providers = append([]configv1.ProviderConfiguration{secretBoxProvider}, ec.Resources[0].Providers...)
	}

	o, err := yaml.Marshal(ec)
	if err != nil {
		return err
	}
	secret.Data[EncryptionProviderConfig] = o
	return nil
}

// removeOldEncryptionKey will either remove legacy aescbc provider config and its key
// or remove the last encryption key in secretbox provider
func removeOldEncryptionKey(secret *v1.Secret) error {
	var ec configv1.EncryptionConfiguration
	err := yaml.Unmarshal(secret.Data[EncryptionProviderConfig], &ec)
	if err != nil {
		return err
	}

	// try to remove legacy aescbc provider if present
	for i, p := range ec.Resources[0].Providers {
		if p.AESCBC != nil {
			// found AESCBC provider configuration, should be removed
			ec.Resources[0].Providers = append(ec.Resources[0].Providers[:i], ec.Resources[0].Providers[i+1:]...)

			o, err := yaml.Marshal(ec)
			if err != nil {
				return err
			}
			secret.Data[EncryptionProviderConfig] = o
			return nil
		}
	}

	// if no aescbc provider present, remove the last key from the secretbox provider
	for _, p := range ec.Resources[0].Providers {
		if p.Secretbox != nil {
			keysCount := len(p.Secretbox.Keys)
			if keysCount > 0 {
				// remove the last key from the array
				p.Secretbox.Keys = p.Secretbox.Keys[:keysCount-1]
			}

			o, err := yaml.Marshal(ec)
			if err != nil {
				return err
			}
			secret.Data[EncryptionProviderConfig] = o
		}
	}
	return nil
}

// rewriteAllSecrets will load all secrets from cluster, add an annotation that marks that it has been rewriten
// and updates them in API
func rewriteAllSecrets(wcClient ctrlclient.Client, ctx context.Context) error {
	var allSecrets v1.SecretList
	err := wcClient.List(ctx, &allSecrets)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format(time.RFC3339)

	for i := range allSecrets.Items {
		if allSecrets.Items[i].Annotations == nil {
			allSecrets.Items[i].Annotations = map[string]string{}
		}
		allSecrets.Items[i].Annotations[AnnotationRewriteTimestamp] = timestamp

		err = wcClient.Update(ctx, &allSecrets.Items[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) countMasterNodesWithLatestConfig(ctx context.Context, wcClient ctrlclient.Client, configShake256Sum string) (bool, error) {
	// get the secret with md5 checksums of the config file
	var shake256Secret v1.Secret
	err := wcClient.Get(ctx,
		ctrlclient.ObjectKey{
			Name:      EncryptionProviderConfigShake256SecretName,
			Namespace: EncryptionProviderConfigShake256SecretNamespace,
		},
		&shake256Secret)
	if apierrors.IsNotFound(err) {
		// secret does not exist yet, not and actual error, lets check next reconciliation loop
		s.logger.Info(fmt.Sprintf("secret %s do not exists yet on the workload cluster", EncryptionProviderConfigShake256SecretName))
		return false, nil
	} else if err != nil {
		return false, err
	}

	// get all master nodes
	var nodes v1.NodeList
	err = wcClient.List(ctx,
		&nodes,
		ctrlclient.MatchingLabels{MasterNodeLabel: ""})
	if err != nil {
		return false, err
	}

	nodeCount := len(nodes.Items)
	if nodeCount != 1 && nodeCount != 3 && nodeCount != 5 {
		err = errors.New("unexpected number of master nodes, cluster is probably in transiting state")
		s.logger.Error(err, fmt.Sprintf("expected 1 or 3 or 5 master nodes but found %d", nodeCount))
		return false, nil
	}

	masterNodeWithLatestConfig := 0
	for _, n := range nodes.Items {
		if v, ok := shake256Secret.Data[n.Name]; ok {
			if string(v) == configShake256Sum {
				// the md5sum matches, this master node has the new config
				masterNodeWithLatestConfig += 1
			}
		}
	}

	if masterNodeWithLatestConfig == nodeCount {
		s.logger.Info(fmt.Sprintf("all masters are running updated encryption provider config (%d/%d are up to date)", masterNodeWithLatestConfig, nodeCount))
		return true, nil
	}

	s.logger.Info(fmt.Sprintf("not all masters are running updated encryption provider config (%d/%d are up to date)", masterNodeWithLatestConfig, nodeCount))
	return false, nil
}

// initNewEncryptionConfigStruct will build struct for the encryption configuration
func initNewEncryptionConfigStruct(provider configv1.ProviderConfiguration) configv1.EncryptionConfiguration {
	return configv1.EncryptionConfiguration{
		Kind:       "EncryptionConfig",
		APIVersion: "v1",
		Resources: []configv1.ResourceConfiguration{
			{
				Resources: []string{"secrets"},
				Providers: []configv1.ProviderConfiguration{
					provider,
					{
						Identity: &configv1.IdentityConfiguration{},
					},
				},
			},
		},
	}
}

// getMaxKeyIndex return biggest index used in key name
func getMaxKeyIndex(keys []configv1.Key) (int, error) {
	index := 0
	for _, k := range keys {
		i, err := getKeyIndex(k)
		if err != nil {
			return 0, err
		}
		if i > index {
			index = i
		}
	}
	return index, nil
}

func getKeyIndex(key configv1.Key) (int, error) {
	keyIndex := strings.TrimPrefix(key.Name, KeyNamePrefix)
	i, err := strconv.Atoi(keyIndex)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func keyName(i int) string {
	return fmt.Sprintf("%s%d", KeyNamePrefix, i)
}

func shake256Sum(buf []byte) string {
	h := make([]byte, 64)
	// Compute a 64-byte hash of buf and put it in h.
	sha3.ShakeSum256(h, buf)
	return fmt.Sprintf("%x\n", h)
}
