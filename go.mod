module github.com/giantswarm/encryption-provider-operator

go 1.16

require (
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/aws/aws-sdk-go v1.40.53 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.5.6 // indirect
	k8s.io/api v0.17.17
	k8s.io/apimachinery v0.17.17
	k8s.io/client-go v0.17.17
	k8s.io/klog v1.0.0
	sigs.k8s.io/cluster-api v0.3.22
	sigs.k8s.io/cluster-api-provider-aws v0.6.8
	sigs.k8s.io/controller-runtime v0.5.14
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
