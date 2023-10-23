# Vault OIDC access token plugin

The OIDC access token plugin is a secret engine for [HashiCorp Vault](https://www.vaultproject.io/) that provides a secure consumption model for [OIDC client credentials flow](https://datatracker.ietf.org/doc/html/rfc6749#section-1.3.4) access tokens. Users can read temporary access tokens of a client without the need to have access to the underlying static credentials of the client.

## Build

Build the plugin binary via

```sh
make build
```

## Giving it a test ride

You can spin up a Vault instance in developer mode and mount the secret engine under `oidc/` by running

```sh
make start
```

You can control the version of Vault that is started with the `VAULT_VERSION` env variable. Defaults to 1.12.7.

Log into Vault with...

```sh
export VAULT_ADDR="http://localhost:8200"
vault login root
```

... and add a client.

```sh
vault write oidc/clients/${client_id} client_id=${client_id} client_secret="${client_secret}" grant_type=client_credentials token_url="https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
Success! Data written to: oidc/clients/....
```

The credentials are validated upon client creation by trying to aquire an access token. So make sure `client_id` and `client_secret` are valid on the provided `token_url`.

Now request an access token

```sh
vault read oidc/accesstoken/$client_id
Key             Value
---             -----
access_token    eyJhbGciOiJSU.......
expiry          2023-10-20T17:06:20.59956+02:00
```
