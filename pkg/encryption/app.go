package encryption

import (
	"context"
	"fmt"

	chartv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/encryption-provider-operator/pkg/project"
)

const (
	configMapName  = "encryption-config-hasher-chart-values"
	chartNamespace = "giantswarm"

	appName      = "encryption-config-hasher"
	appNamespace = "kube-system"

	encryptionConfigHasherVersion = "0.1.1"
)

var (
	chartURL = fmt.Sprintf("https://giantswarm.github.io/giantswarm-playground-catalog/encryption-config-hasher-%s.tgz", encryptionConfigHasherVersion)
)

func (s *Service) deployEncryptionProviderHasherApp(ctx context.Context, wcClient ctrlclient.Client) error {
	cm := buildConfigMapValues(s.registryDomain)

	err := wcClient.Create(ctx, cm)
	if apierrors.IsAlreadyExists(err) {
		// fall through
	} else if err != nil {
		return err
	}

	chart := buildAppChart()

	err = wcClient.Create(ctx, chart)

	if apierrors.IsAlreadyExists(err) {
		// fall through
	} else if err != nil {
		return err
	}

	return nil
}

func (s *Service) deleteEncryptionProviderHasherApp(ctx context.Context, wcClient ctrlclient.Client) error {
	cm := buildConfigMapValues(s.registryDomain)

	err := wcClient.Delete(ctx, cm)
	if apierrors.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return err
	}

	chart := buildAppChart()

	err = wcClient.Delete(ctx, chart)
	if apierrors.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return err
	}

	return nil
}

type Values struct {
	Registry Registry `yaml:"registry"`
}
type Registry struct {
	Domain string `yaml:"domain"`
}

func buildConfigMapValues(registryDomain string) *v1.ConfigMap {
	val := Values{
		Registry: Registry{
			Domain: registryDomain,
		},
	}
	values, _ := yaml.Marshal(&val)

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: chartNamespace,
			Labels: map[string]string{
				label.AppKubernetesName: "encryption-config-hasher",
			},
		},
		Data: map[string]string{
			"values": string(values),
		},
	}
	return cm
}

func buildAppChart() *chartv1.Chart {
	c := &chartv1.Chart{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: chartNamespace,
			Labels: map[string]string{
				label.AppKubernetesName:    appName,
				label.ManagedBy:            project.Name(),
				label.ChartOperatorVersion: "1.0.0",
			},
			Annotations: map[string]string{
				annotation.ChartOperatorForceHelmUpgrade: "true",
			},
		},
		Spec: chartv1.ChartSpec{
			Name:      appName,
			Namespace: appNamespace,
			Config: chartv1.ChartSpecConfig{
				ConfigMap: chartv1.ChartSpecConfigConfigMap{
					Name:      configMapName,
					Namespace: chartNamespace,
				},
			},
			TarballURL: chartURL,
			Version:    encryptionConfigHasherVersion,
		},
	}
	return c
}
