package encryption

type Data struct {
	Keys []Key
}

type Key struct {
	EncryptionKey string
	Name          string
	Type          string
}

const encryptionConfigTemplate = `kind: EncryptionConfig
apiVersion: v1
resources:
  - resources:
    - secrets
    providers:
    {{- range $key := .Keys }}
    - {{$key.Type}}:
        keys:
        - name: {{$key.Name}}
          secret: {{$key.EncryptionKey}}
    {{- end }}
    - identity: {}
`
