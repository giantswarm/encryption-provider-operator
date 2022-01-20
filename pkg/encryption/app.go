package encryption

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/label"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//chartv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
)

const (
	configMapName  = "encryption-config-hasher-chart-values"
	chartNamespace = "ginatswarm"
)

func (s *Service) DeployEncryptionProviderHasherApp(ctx context.Context) error {
	cm := buildConfigMapValues(s.registryDomain)

	err := s.ctrlClient.Create(ctx, &cm)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteEncryptionProviderHasherApp() error {

	return nil
}

func buildConfigMapValues(registryDomain string) v1.ConfigMap {
	tmpl := `|
registry:
  domain: %s`
	values := fmt.Sprintf(tmpl, registryDomain)

	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: chartNamespace,
			Labels: map[string]string{
				label.AppKubernetesName: "encryption-config-hasher",
			},
		},
		Data: map[string]string{
			"values": values,
		},
	}

	return cm
}

//func buildChart() chartv1.Chart {
//
//}
