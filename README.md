[![CircleCI](https://circleci.com/gh/giantswarm/encryption-provider-operator.svg?style=shield)](https://circleci.com/gh/giantswarm/encryption-provider-operator)

# encryption-provider-operator

encryption-provider-operator is creating and updating encryption config for k8s secret encryption of secret in etcd

simplified process of key rotation
* trigger new keyrotation  -> either via annotation or after some period
* new encryption config file is generated with old and new key, the new key on the first position
* install encryption config hasher on the cluster and calculate hashes
* operator waits until all nodes have the hash of the config that is equal to what it sees in the MC
* operator will recreate all secrets
* operator will update the encryption config and remove the old key
the * last step is to roll all master nodes again but it's not required or watched by the controller
