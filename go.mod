module github.com/giantswarm/encryption-provider-operator

go 1.16

require (
	github.com/giantswarm/apiextensions-application v0.3.0
	github.com/giantswarm/k8smetadata v0.8.0
	github.com/go-logr/logr v1.2.2
	github.com/google/go-cmp v0.5.7
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.2
	k8s.io/apimachinery v0.23.2
	k8s.io/client-go v0.23.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/cluster-api v1.0.2
	sigs.k8s.io/controller-runtime v0.11.0
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
