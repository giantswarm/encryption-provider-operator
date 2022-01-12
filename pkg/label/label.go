package label

const (
	// Cluster label is a new style label for ClusterID
	Cluster = "giantswarm.io/cluster"
	// ManagedBy label denotes which operator manages corresponding resource.
	ManagedBy = "giantswarm.io/managed-by"
	// RandomKey label specifies type of a secret that is used for guest
	// cluster.
	RandomKey = "giantswarm.io/randomkey"
	// RandomKeyTypeEncryption is a type of randomkey secret used for guest
	// cluster.
	RandomKeyTypeEncryption = "encryption"
)
