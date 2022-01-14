module github.com/giantswarm/encryption-provider-operator

go 1.16

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/giantswarm/k8smetadata v0.7.1
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/onsi/gomega v1.14.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602 // indirect
	k8s.io/api v0.17.17
	k8s.io/apiextensions-apiserver v0.17.17 // indirect
	k8s.io/apimachinery v0.17.17
	k8s.io/client-go v0.17.17
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.1.0 // indirect
	k8s.io/utils v0.0.0-20210709001253-0e1f9d693477 // indirect
	sigs.k8s.io/cluster-api v0.3.22
	sigs.k8s.io/controller-runtime v0.5.14
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
