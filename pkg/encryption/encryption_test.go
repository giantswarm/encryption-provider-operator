package encryption

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"

	configv1 "github.com/giantswarm/encryption-provider-operator/pkg/config"
)

func Test_removeOldEncryptionKey(t *testing.T) {
	testCases := []struct {
		name           string
		secret         v1.Secret
		expectedSecret v1.Secret
	}{
		{
			name: "case 0:  remove old aescbc key",
			secret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey1
  - aescbc:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
			expectedSecret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
		},
		{
			name: "case 1:  remove old secretbox key",
			secret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey1
      - name: key0
        secret: testkey0
  - identity: {}
`)},
			},
			expectedSecret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
		},
		{
			name: "case 2: only 1 key in config, nothing should be removed",
			secret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
			expectedSecret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := removeOldEncryptionKey(&tc.secret)
			if err != nil {
				t.Fatalf(" %s : failed to remove old encryption key %s", tc.name, err)
			}

			if !reflect.DeepEqual(tc.secret, tc.expectedSecret) {
				t.Fatalf("%s : secrets are not equal %s", tc.name, cmp.Diff(string(tc.secret.Data[EncryptionProviderConfig]), string(tc.expectedSecret.Data[EncryptionProviderConfig])))
			}
		})
	}
}

func Test_addNewEncryptionKey(t *testing.T) {
	testCases := []struct {
		name           string
		secret         v1.Secret
		expectedSecret v1.Secret
	}{
		{
			name: "case 0: add new key to config with secretbox provider",
			secret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
			expectedSecret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key2
        secret: testkey0
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
		},
		{
			name: "case 1: add new key to config with aescbc provider",
			secret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - aescbc:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
			expectedSecret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey0
  - aescbc:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := addNewEncryptionKey(&tc.secret, "testkey0")
			if err != nil {
				t.Fatalf(" %s : failed to add new encryption key to config %s", tc.name, err)
			}

			if !reflect.DeepEqual(tc.secret, tc.expectedSecret) {
				fmt.Printf("%s\n%s\n", string(tc.secret.Data[EncryptionProviderConfig]), string(tc.expectedSecret.Data[EncryptionProviderConfig]))
				t.Fatalf("%s : secrets are not equal %s", tc.name, cmp.Diff(string(tc.secret.Data[EncryptionProviderConfig]), string(tc.expectedSecret.Data[EncryptionProviderConfig])))
			}
		})
	}
}

func Test_getMaxKeyIndex(t *testing.T) {
	testCases := []struct {
		name          string
		secret        v1.Secret
		expectedIndex int
	}{
		{
			name: "case 0: single key",
			secret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
			expectedIndex: 1,
		},
		{
			name: "case 1: multiple keys",
			secret: v1.Secret{
				Data: map[string][]byte{EncryptionProviderConfig: []byte(`kind: EncryptionConfiguration
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - secretbox:
      keys:
      - name: key4
        secret: testkey1
      - name: key3
        secret: testkey1
      - name: key2
        secret: testkey1
      - name: key1
        secret: testkey1
  - identity: {}
`)},
			},
			expectedIndex: 4,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var ec configv1.EncryptionConfiguration
			err := yaml.Unmarshal(tc.secret.Data[EncryptionProviderConfig], &ec)
			if err != nil {
				t.Fatalf("%s : failed to unmarshal struct %s", tc.name, err)
			}

			i, err := getMaxKeyIndex(ec.Resources[0].Providers[0].Secretbox.Keys)
			if err != nil {
				t.Fatalf("%s : failed get last key index %s", tc.name, err)
			}

			if i != tc.expectedIndex {
				t.Fatalf("%s : expected index %d but got %d", tc.name, tc.expectedIndex, i)
			}
		})
	}
}
