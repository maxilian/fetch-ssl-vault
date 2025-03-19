## Fetch SSL from Hashicorp Vault

This code is assisted by ChatGPT so I'd like to make it open source. 

The idea behind this code is I need a script that can fetch SSL certificate which is stored on Vault and then store the responses into specific path on the disk periodically whether the OS you use is Unix-like or Windows.


### How to use:

- The certificate must include private key stored in Vault secret and it's named with tls.crt for the cert and tls.key for the private key.
- You must fill the `.env` and provide approle `role_id` and `secret_id`.

```
mv .env.dev .env
go mod tidy
go run main.go
```