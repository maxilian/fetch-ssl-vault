## Fetch SSL from Hashicorp Vault

This code is assisted with ChatGPT so I make it open source. The idea behind this code is I need a script that can fetch SSL certificate (which is stored in Vault) and store the responses to specific path on the disk whether the OS you use is Unix-like or Windows.


### How to use:

- The certificate must include private key stored in Vault and it's named with tls.crt for the cert and tls.key for the private key.
- You must provide approle `role_id` and `secret_id` in .env file.

```
mv .env.dev .env
go mod tidy
go run main.go
```