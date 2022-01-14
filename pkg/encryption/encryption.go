package encryption

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/encryption-provider-operator/pkg/key"
	"github.com/giantswarm/encryption-provider-operator/pkg/label"
	"github.com/giantswarm/encryption-provider-operator/pkg/project"
)

const (
	EncryptionProviderConfig = "encryption"

	// LegacyEncryptionProviderType is old encryption provider used  by legacy giantswarm product
	LegacyEncryptionProviderType = "aescbc"

	// DefaultEncryptionProviderType value is used for new keys
	// https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#providers
	DefaultEncryptionProviderType = "secretbox"

	// Poly1305KeyLength represents the 32 bytes length for Poly1305
	// padding encryption key.
	Poly1305KeyLength = 32
)

type Config struct {
	Cluster                  *capi.Cluster
	DefaultKeyRotationPeriod time.Duration

	CtrlClient ctrlclient.Client
	Logger     logr.Logger
}

type Service struct {
	cluster *capi.Cluster

	defaultKeyRotationPeriod time.Duration

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
	if c.Logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	s := &Service{
		cluster:                  c.Cluster,
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
		err = s.keyRotation(ctx)
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

	var encryptionKey string
	var encryptionProviderType string

	if apierrors.IsNotFound(err) {
		// no old key found, lets generate a new one
		newKey, err := newRandomKey(Poly1305KeyLength)
		if err != nil {
			s.logger.Error(err, "failed to generate new random key for encryption")
			return err
		}
		encryptionKey = newKey
		encryptionProviderType = DefaultEncryptionProviderType

		s.logger.Info("generated a new encryption key for Poly1305 encryption provider")

	} else if err != nil {
		s.logger.Error(err, "failed to get old encryption provider key secret")
		return err
	} else {
		// there is an old encryption key so lets reuse it to avoid breaking cluster
		if k, ok := oldEncryptionSecret.Data["encryption"]; !ok {
			s.logger.Error(err, "failed to get encryption key from secret")
			return err
		} else {
			encryptionKey = string(k)
			// the old encryption key are using old less secure type
			// https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#providers
			encryptionProviderType = LegacyEncryptionProviderType
			s.logger.Info("fetched and migrated AESCBC encryption key from legacy product")

		}
	}

	secretData := map[string]string{}
	// render template for the config
	{
		templateData := Data{
			Keys: []Key{
				{
					Name:          "key1",
					EncryptionKey: encryptionKey,
					Type:          encryptionProviderType,
				},
			},
		}
		tmpl, err := template.New("tmpl").Parse(encryptionConfigTemplate)
		if err != nil {
			return err
		}
		var buff bytes.Buffer

		err = tmpl.Execute(&buff, templateData)
		if err != nil {
			return err
		}

		secretData[EncryptionProviderConfig] = buff.String()
	}

	encryptionProviderSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.SecretName(clusterName),
			Namespace: s.cluster.Namespace,
			Labels: map[string]string{
				label.Cluster:        clusterName,
				label.ManagedBy:      project.Name(),
				label.RandomKey:      label.RandomKeyTypeEncryption,
				key.ClusterNameLabel: clusterName,
			},
		},
		StringData: secretData,
	}

	err = s.ctrlClient.Create(ctx, encryptionProviderSecret)
	if err != nil {
		s.logger.Error(err, "failed to create encryption provider secret")
		return err
	}

	s.logger.Info("created a new encryption provider config secret")

	return nil
}

func (s *Service) keyRotation(ctx context.Context) error {
	// TODO implement the magic

	return nil
}

func newRandomKey(length int) (string, error) {
	randomKey := make([]byte, length)

	_, err := rand.Read(randomKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(randomKey), nil
}
