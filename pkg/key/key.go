package key

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ClusterNameLabel        = "cluster.x-k8s.io/cluster-name"
	ClusterWatchFilterLabel = "cluster.x-k8s.io/watch-filter"

	FinalizerName = "encryption-provider-operator.finalizers.giantswarm.io"
)

func GetClusterIDFromLabels(t v1.ObjectMeta) string {
	return t.GetLabels()[ClusterNameLabel]
}

func SecretName(clusterName string) string {
	return fmt.Sprintf("%s-encryption-provider-config", clusterName)
}
