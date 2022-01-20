package key

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterNameLabel = "cluster.x-k8s.io/cluster-name"

	FinalizerName = "encryption-provider-operator.finalizers.giantswarm.io"
)

func GetClusterIDFromLabels(t metav1.ObjectMeta) string {
	return t.GetLabels()[ClusterNameLabel]
}

func SecretName(clusterName string) string {
	return fmt.Sprintf("%s-encryption-provider-config", clusterName)
}

// GetWCK8sClient will return workload cluster k8s controller-runtime client
func GetWCK8sClient(ctx context.Context, ctrlClient client.Client, clusterName string, clusterNamespace string) (client.Client, error) {
	var err error

	if _, err := os.Stat(tempKubeconfigFileName(clusterName)); err == nil {
		// kubeconfig file already exists, no need to fetch and write again

	} else if os.IsNotExist(err) {
		// kubeconfig dont exists we need to fetch it and write to file
		var secret corev1.Secret
		{
			err = ctrlClient.Get(ctx, client.ObjectKey{
				Name:      fmt.Sprintf("%s-kubeconfig", clusterName),
				Namespace: clusterNamespace,
			},
				&secret)

			if err != nil {
				return nil, err
			}
		}
		err = ioutil.WriteFile(tempKubeconfigFileName(clusterName), secret.Data["value"], 0600)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", tempKubeconfigFileName(clusterName))
	if err != nil {
		return nil, err
	}

	wcClient, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}

	return wcClient, nil
}

// CleanWCK8sKubeconfig will clean old kubeconfig file to avoid issue when cluster is recreated with same ID
func CleanWCK8sKubeconfig(clusterName string) error {
	err := os.Remove(tempKubeconfigFileName(clusterName))
	if os.IsNotExist(err) {
		// we ignore if the file is already deleted
	} else if err != nil {
		return err
	}

	return nil
}

func tempKubeconfigFileName(clusterName string) string {
	return fmt.Sprintf("/tmp/kubeconfig-%s", clusterName)
}
