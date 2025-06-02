# Secret Management with Vault

This project stores sensitive configuration such as database and object storage
passwords in [HashiCorp Vault](https://www.vaultproject.io/). The main `values.yaml`
file does not contain any credentials. Instead, a helper script retrieves the
required secrets and generates a separate values file.

## Preparing Vault

1. Install Vault and start a dev server or connect to your existing instance.
2. Enable the KV engine and create the secrets used by the Helm chart:

   ```bash
   vault secrets enable -path=secret kv
   vault kv put secret/twitter/postgres user=myuser password=mypassword
   vault kv put secret/twitter/minio accessKey=minio secretKey=minio123
   ```

## Generating the values file

Run the helper script which reads the secrets from Vault and writes
`helm-chart/values.secrets.yaml`:

```bash
./scripts/fetch-vault-secrets.sh
```

This file contains only the sensitive keys and should **not** be committed to
the repository. Deploy the chart by passing it alongside `values.yaml`:

```bash
helm upgrade --install twitter-clone ./helm-chart \
  -f helm-chart/values.yaml -f helm-chart/values.secrets.yaml
```

The application will then start with the credentials sourced securely from Vault.
