package project

var (
	description = "The encryption-provider-operator manages encryption provider configs to encrypt k8s secrete data in etcd."
	gitSHA      = "n/a"
	name        = "encryption-provider-operator"
	source      = "https://github.com/giantswarm/encryption-provider-operator"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}