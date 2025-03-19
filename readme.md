## Fetch SSL from Hashicorp Vault

This code is assisted by ChatGPT so I'd like to make it an open source project. 

The idea behind this code is I need a script that can fetch SSL certificate which is stored on Vault periodically and then save the responses into specific path on the disk whether the OS you use is Unix-like or Windows.


### How to use:

- The certificate must include private key stored in Vault secret and it's named with tls.crt for the cert and tls.key for the private key.
- You must fill the `.env` and provide approle `role_id` and `secret_id`.
- The schedule is in crontab format, you can use something like `0 1 * * 1` for example to run every Monday at 1AM UTC.
- For Windows usage, you must escape slash for cert path if you use absolute path, for example `C:\\nginx\\conf\\ssl`.

```
mv .env.dev .env
go mod tidy
go run main.go
```