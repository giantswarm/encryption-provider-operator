package encryption

import (
	"context"
	"fmt"

	chartv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
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

	encryptionConfigHasherVersion = "0.3.0"
)

func chartURL(appCatalog string) string {
	return fmt.Sprintf("https://giantswarm.github.io/%s/encryption-config-hasher-%s.tgz", appCatalog, encryptionConfigHasherVersion)
}

func (s *Service) deployEncryptionProviderHasherApp(ctx context.Context, wcClient ctrlclient.Client) error {
	cm := buildConfigMapValues(s.registryDomain)

	err := wcClient.Create(ctx, cm)
	if apierrors.IsAlreadyExists(err) {
		// update chart of it already exists
		err = wcClient.Get(ctx, ctrlclient.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, cm)
		if err != nil {
			return microerror.Mask(err)
		}

		cm.Data = configMapData(s.registryDomain)

		err = wcClient.Update(ctx, cm)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}

	chart := buildAppChart(s.appCatalog)

	err = wcClient.Create(ctx, chart)

	if apierrors.IsAlreadyExists(err) {
		// update chart of it already exists
		err = wcClient.Get(ctx, ctrlclient.ObjectKey{Name: chart.Name, Namespace: chart.Namespace}, chart)
		if err != nil {
			return microerror.Mask(err)
		}
		chart.Spec = chartSpec(s.appCatalog)

		err = wcClient.Update(ctx, chart)
		if err != nil {
			return microerror.Mask(err)
		} else {
			s.logger.Info(fmt.Sprintf("updated '%s' app in workload cluster", chart.Name))
		}

	} else if err != nil {
		return microerror.Mask(err)
	} else {
		s.logger.Info(fmt.Sprintf("deployed '%s' app to workload cluster", chart.Name))
	}

	return nil
}

func (s *Service) deleteEncryptionProviderHasherApp(ctx context.Context, wcClient ctrlclient.Client) error {
	cm := buildConfigMapValues(s.registryDomain)

	err := wcClient.Delete(ctx, cm)
	if apierrors.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return microerror.Mask(err)
	}

	chart := buildAppChart(s.appCatalog)

	err = wcClient.Delete(ctx, chart)
	if apierrors.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return microerror.Mask(err)
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
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: chartNamespace,
			Labels: map[string]string{
				label.AppKubernetesName: "encryption-config-hasher",
			},
		},
		Data: configMapData(registryDomain),
	}
	return cm
}

func configMapData(registryDomain string) map[string]string {
	val := Values{
		Registry: Registry{
			Domain: registryDomain,
		},
	}
	values, _ := yaml.Marshal(&val)

	return map[string]string{"values": string(values)}
}

func buildAppChart(appCatalog string) *chartv1.Chart {
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
		Spec: chartSpec(appCatalog),
	}
	return c
}

func chartSpec(appCatalog string) chartv1.ChartSpec {
	return chartv1.ChartSpec{
		Name:      appName,
		Namespace: appNamespace,
		Config: chartv1.ChartSpecConfig{
			ConfigMap: chartv1.ChartSpecConfigConfigMap{
				Name:      configMapName,
				Namespace: chartNamespace,
			},
		},
		TarballURL: chartURL(appCatalog),
		Version:    encryptionConfigHasherVersion,
	}
}
