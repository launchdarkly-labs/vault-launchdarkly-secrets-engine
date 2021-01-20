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

2. Enable the launchdarkly secrets engine:

    ```text
    $ vault secrets enable -path="launchdarkly" -plugin-name="vault-launchdarkly-secret-engine" plugin
    Success! Enabled the vault-launchdarkly-secret-engine plugin at: launchdarkly/
    ```

    By default, the secrets engine will mount at the name of the engine. To
    enable the secrets engine at a different path, use the `-path` argument.

3. Configure the backend with user credentials that will be able to interact with the LaunchDarkly API and create tokens. Additional information about Token Creation is below.

    ```text
    $ vault write launchdarkly/config accessToken="123456789"
    Success! Data written to: launchdarkly/config
    ```

4. There is optional configuration parameters `ttl` and `max_ttl` that if set will override the default system TTL for tokens issued from this Secret Engine.

### Generating Token
You can create a token by going to [Authorization](https://app.launchdarkly.com/settings/authorization/tokens/new) in your Dashboard.

```
[
  {
    "effect": "allow",
    "resources": [
      "proj/*:env/*"
    ],
    "actions": [
      "updateApiKey",
      "updateMobileKey"
    ]
	},
  {
    "effect": "allow",
    "resources": [
      "code-reference-repository/*"
    ],
    "actions": [
      "*"
    ]
	},
    {
    "effect": "allow",
    "resources": [
      "relay-proxy-config/*"
    ],
    "actions": [
      "*"
    ]
	},
      {
    "effect": "allow",
    "resources": [
      "service-token/*"
    ],
    "actions": [
      "*"
    ]
	}
]
```
### Usage

After the secrets engine is configured and a user/machine has a Vault token with the proper permission, it can generate tokens.

1. Generate a new LaunchDarkly token by reading from the  `launchdarkly/role/<custom-role-key>` endpoint. Each read will generate a new token and associated TTL:

    ```text
    $ vault read launchdarkly/role/<custom-role-key>
    Key                Value
    ---                -----
    lease_id           launchdarkly/role/api-writer/DIQhrlO5XoLAQrBKUowGXZ8h
    lease_duration     768h
    lease_renewable    true
    token              api-12345
    ```

2. To read SDK keys for an Environment `launchdarkly/project/<project-key>/<environment-key>`. The keys are not returned as a Secret unlike API Tokens, they are long-lived values that are the same for all clients:

    ```text
    $ vault read launchdarkly/project/test-project/development
    Key         Value
    ---         -----
    clientId    aaaabde7df99999900000
    mobile      mob-826b0530-9999-0000-aed4-12345
    sdk         sdk-65f59771-0000-9999-b567-12345
    ```

3. To reset a SDK or Mobile key you can read from: `vault read launchdarkly/project/<project-key>/<environment-key>/reset/sdk` where the final string can be `sdk` or `mobile`.

Paths:
```
info - Returns build information the Secret Engine version.
config - Configuration for the plugin.
role - Generates tokens for associated LaunchDarkly Custom Roles.
relay - After writing a policy to Vault storage, it will generate tokens for that policy.
coderefs - Generate short-lived tokens to push over Code References.
```

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
vault write sys/plugins/catalog/secret/vault-launchdarkly-secrets-engine sha_256="$SHASUM" command="vault-launchdarkly-secrets-engine"
vault secrets enable -path="launchdarkly" -plugin-name="vault-launchdarkly-secrets-engine" plugin
vault write launchdarkly/config access_token="123456789"
```
