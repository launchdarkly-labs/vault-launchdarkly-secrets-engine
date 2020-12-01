# vault-launchdarkly-secret-engine

This plugin will allow you to create a secret backend that will use the LaunchDarkly API to generate dynamic LaunchDarkly tokens.  Usage can be restricted using the highly customizable Vault ACL system.

Forked from: https://github.com/nytimes/vault-fastly-secret-engine

### Setup

Most secrets engines must be configured in advance before they can perform their
functions. These steps are usually completed by an operator or configuration
management tool.

1. Register the plugin with the catalog

    ```text
    $ SHASUM=$(shasum -a 256 vault-launchdarkly-secret-engine | cut -d " " -f1)
    $ vault write sys/plugins/catalog/vault-launchdarkly-secret-engine sha_256="$SHASUM" command="vault-launchdarkly-secret-engine"
    Success! Data written to: sys/plugins/catalog/vault-launchdarkly-secret-engine
    ```

1. Enable the launchdarkly secrets engine:

    ```text
    $ vault secrets enable -path="launchdarkly" -plugin-name="vault-launchdarkly-secret-engine" plugin
    Success! Enabled the vault-launchdarkly-secret-engine plugin at: launchdarkly/
    ```

    By default, the secrets engine will mount at the name of the engine. To
    enable the secrets engine at a different path, use the `-path` argument.

1. Configure the backend with user credentials that will be able to interact with the LaunchDarkly API and create tokens.

    ```text
    $ vault write launchdarkly/config accessToken="123456789"
    Success! Data written to: launchdarkly/config
    ```

### Usage

After the secrets engine is configured and a user/machine has a Vault token with the proper permission, it can generate tokens.

1. Generate a new LaunchDarkly token by reading from the  `launchdarkly/role/<custom-role-key>` endpoint:

    ```text
    $ vault read launchdarkly/role/<custom-role-key>"
    Key      Value
    ---      -----
    token    api-123456789
    ```

2. To read SDK keys for an Environment from the `launchdarkly/project/<project-key>/<environment-key>`:

    ```text
    $ vault read launchdarkly/project/test-project/development
    Key         Value
    ---         -----
    clientId    aaaabde7df99999900000
    mobile      mob-826b0530-9999-0000-aed4-12345
    sdk         sdk-65f59771-0000-9999-b567-12345
    ```

3. To reset a SDK or Mobile key you can read from: `vault read launchdarkly/project/<project-key>/<environment-key>/reset/sdk` where the final string can be `sdk` or `mobile`.

## Local Development

### Build the code

```bash
goreleaser build --snapshot --rm-dist
docker build -t vault-plugin .
docker run --cap-add=IPC_LOCK -e 'VAULT_DEV_ROOT_TOKEN_ID=myroot' -e 'VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:1234' -p 1234:1234 vault-plugin
```

### Configure the local vault

In a second terminal window...

```bash
export VAULT_ADDR='http://0.0.0.0:1234'
vault login myroot
SHASUM=$(shasum -a 256 dist/vault-launchdarkly-secrets-engine_linux_amd64/vault-launchdarkly-secrets-engine | cut -d " " -f1)
vault write sys/plugins/catalog/secret/vault-launchdarkly-secret-engine   sha_256="$SHASUM"   command="vault-launchdarkly-secrets-engine"
vault secrets enable -path="launchdarkly" -plugin-name="vault-launchdarkly-secrets-engine" plugin
vault write launchdarkly/config accessToken="123456789"
```
