package key

import (
	"context"
	"fmt"
	"os"

	chartv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FinalizerName = "encryption-provider-operator.finalizers.giantswarm.io"

	MasterNodeLabel = "node-role.kubernetes.io/master"
)

func SecretName(clusterName string) string {
	return fmt.Sprintf("%s-encryption-provider-config", clusterName)
}

// GetWCK8sClient will return workload cluster k8s controller-runtime client
func GetWCK8sClient(ctx context.Context, ctrlClient client.Client, clusterName string, clusterNamespace string) (client.Client, error) {
	var err error
	var kubeconfig []byte
	var secret corev1.Secret
	{
		err = ctrlClient.Get(ctx, client.ObjectKey{
			Name:      fmt.Sprintf("%s-kubeconfig", clusterName),
			Namespace: clusterNamespace,
		},
			&secret)

		if apierrors.IsNotFound(err) {
			// in legacy gs the kubeconfig is in namespace equal to cluster ID
			err = ctrlClient.Get(ctx, client.ObjectKey{
				Name:      fmt.Sprintf("%s-kubeconfig", clusterName),
				Namespace: clusterName,
			},
				&secret)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			kubeconfig = secret.Data["kubeConfig"]
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			kubeconfig = secret.Data["value"]
		}
	}
	err = os.WriteFile(tempKubeconfigFileName(clusterName), kubeconfig, 0600)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", tempKubeconfigFileName(clusterName))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	scheme := runtime.NewScheme()
	_ = chartv1.AddToScheme(scheme)
	_ = clientgoscheme.AddToScheme(scheme)

	wcClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return wcClient, nil
}

// CleanWCK8sKubeconfig will clean old kubeconfig file to avoid issue when cluster is recreated with same ID
func CleanWCK8sKubeconfig(clusterName string) error {
	err := os.Remove(tempKubeconfigFileName(clusterName))
	if os.IsNotExist(err) {
		// we ignore if the file is already deleted
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func tempKubeconfigFileName(clusterName string) string {
	return fmt.Sprintf("/tmp/kubeconfig-%s", clusterName)
}
