# Vault
Данная практическая работа посвящена изучению базовому взаимодействию с
инструментом управления секретами [hashicorp vault][vault].

## Vagrant
Для работы с [vault][] воспользуемся следующим `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "node" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "node"
    c.vm.network "forwarded_port", guest: 8200, host: 8200
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq docker.io
      usermod -a -G docker vagrant
      docker run -it --rm -d -p 8200:8200 --name vault vault:1.13.3
      docker run --rm vault:1.13.3 cat /bin/vault > /usr/local/bin/vault
      chmod +x /usr/local/bin/vault
      curl -L https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-linux-amd64.tar.gz \
        | tar xvz --strip-components=1 -C /usr/local/bin age/age age/age-keygen
      curl -L https://github.com/getsops/sops/releases/download/v3.8.1/sops-v3.8.1.linux.amd64 \
        -o /usr/local/bin/sops && chmod +x /usr/local/bin/sops
    SHELL
  end
end
```

После запуска виртуальной машины необходимо авторизоваться в запущенном vault
server, который запущен в `dev` режиме, так что в его логе можно увидеть токен
для авторизации:
```console
$ docker logs vault | tail
    $ export VAULT_ADDR='http://0.0.0.0:8200'

The unseal key and root token are displayed below in case you want to
seal/unseal the Vault or re-authenticate.

Unseal Key: bo1HYFucRR7b7tfopTOjV4y2vW9xsmU8b8FzKvcOdRo=
Root Token: hvs.FIynHVPht3NxZGgotnzzjmIF

Development mode should NOT be used in production installations!
```

В дальнейшем для работы нам потребуется авторизоваться указав адрес в переменной
`VAULT_ADDR` и токен в команде `vault login`:
```console
$ export VAULT_ADDR='http://0.0.0.0:8200'
$ vault login
Token (will be hidden):
Success! You are now authenticated. The token information displayed below
is already stored in the token helper. You do NOT need to run "vault login"
again. Future Vault requests will automatically use this token.

Key                  Value
---                  -----
token                hvs.FIynHVPht3NxZGgotnzzjmIF
token_accessor       50RuiugSTnGr4VHZrTUgR71l
token_duration       ∞
token_renewable      false
token_policies       ["root"]
identity_policies    []
policies             ["root"]
```

## KV
[Vault][] обладает довольно широким функционалом по работе с секретами, в нем есть
различные движки для разных типов чувствительной информации. Самым простым из
них является [key-value][kv-engine] хранилище, который активирован по-умолчанию и
доступен по пути `secret/`. Для доступа можно воспользоваться командой `vault kv`:
```console
$ vault kv put secret/test somekey=secretvalue
== Secret Path ==
secret/data/test

======= Metadata =======
Key                Value
---                -----
created_time       2023-11-20T20:48:42.050403325Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1
$ vault kv list secret/
Keys
----
test

$ vault kv get secret/test
== Secret Path ==
secret/data/test

======= Metadata =======
Key                Value
---                -----
created_time       2023-11-20T20:48:42.050403325Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

===== Data =====
Key        Value
---        -----
somekey    secretvalue

$ vault kv get -format=json secret/test
{
  "request_id": "751ff986-5a18-c694-4604-68fffd946d36",
  "lease_id": "",
  "lease_duration": 0,
  "renewable": false,
  "data": {
    "data": {
      "somekey": "secretvalue"
    },
    "metadata": {
      "created_time": "2023-11-20T20:48:42.050403325Z",
      "custom_metadata": null,
      "deletion_time": "",
      "destroyed": false,
      "version": 1
    }
  },
  "warnings": null
}

$ vault kv get -field=somekey secret/test
secretvalue

```
Как видно с помощью утилиты `vault` можно получить доступ к хранилищу, а также
получить вывод в различных форматах, что удобно использовать в скриптах или
CI/CD пайплайнах.

Также данное хранилище имеет версионирование и дает возможность получить данные
о секретах старых версий:
```console
$ vault kv put secret/test somekey=anothervalue
== Secret Path ==
secret/data/test

======= Metadata =======
Key                Value
---                -----
created_time       2023-11-20T20:57:28.206932401Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            2
$ vault kv get -field=somekey secret/test
anothervalue
$ vault kv get -field=somekey -version 1 secret/test
secretvalue
```

## Certs
Другим полезным движком для хранения секретов является [PKI][pki-engine], который
позволяет создать собственный удостоверяющий центр сертификации и выписывать
сертификаты. Данный движок не активирован по-умолчанию, так что необходимо
выполнить команду:
```console
$ vault secrets enable pki
Success! Enabled the pki secrets engine at: pki/
```

После чего данный движок будет доступен по пути `pki/` и можно будет сгенерировать
корневой сертификат командой:
```console
$ vault write -field=certificate pki/root/generate/internal \
     common_name="example.com" \
     issuer_name="root-2023" \
     ttl=87600
-----BEGIN CERTIFICATE-----
MIIDNTCCAh2gAwIBAgIUPr8UTRzxlri/a3K+O3fpwC0GhEQwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMjMxMTIwMjExMDExWhcNMjMx
MTIxMjEzMDQxWjAWMRQwEgYDVQQDEwtleGFtcGxlLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAMMquA00dkbKRJwcy3R+QbXuVS/b59S3GTMkQsUj
wc5IIJ70BJQDAdQu0x4TU5k6andV33R8UgIMVNTzDePdddlNLJg9W8EDcvIwPg/Z
UReSCw+fvm+C8wcKqUt3PxJ6z05w7JKB54/sDywCnwX4LRg4XDHEFnvOlvVGmTqO
iwdOTel5hjS8fDRvF8M4HHZEqgUii1+YBhm9DlPpHolGFSVyZ2iA/0YBbgS5Sb+/
gKUzXImKOjNito/X9O+TCR9HqxoiMG7HjC4r3LMy+AZ8LuvVJpp7mor1FRiB0iZ0
z7ZbEYr4kEmTJp75CWM2u+HKGlQ3u/ugSndPhweGJnGzZisCAwEAAaN7MHkwDgYD
VR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFK5aQzjnvFcC
dOl/VOacaZpKsw9PMB8GA1UdIwQYMBaAFK5aQzjnvFcCdOl/VOacaZpKsw9PMBYG
A1UdEQQPMA2CC2V4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQC5o7IF6FsX
j4EA4KI1lhr6Kv60kJhaI5SaCIxc9o8KTShBkGRQD02Fuk9zxdJlJCnyttIDvzLk
pW/fQYhFPgjwIioH1avVRfd5Hh3vkLDxsLMN+9lyy6uS9KxDrWz/cdZUGfMOoLmS
6TAidhy4Z06wlk9IKMTbO3O+LBIr7oIpxH2laOAIwdTClnRwFZCwGkkPsp08l3n5
ZKg9ohy//H1pXx0grn7GXhlONPUUjP8rDZfc/Z5Eo5KlBRCIcW4WcdsPdH9HFxQv
O9r9M0S9iPyzlbZAdt+RhvM8mNU+CGpa1OxE5i2Y0eAITft8jfoA+MMyDC0NmF2J
W/JERPwFQdmK
-----END CERTIFICATE-----
```
После чего его можно увидеть в движке командой:
```console
$ vault list pki/issuers/
Keys
----
79e5ab6e-8030-d428-5c3a-5f7a687d9f7c
$ vault read -field=certificate pki/issuer/79e5ab6e-8030-d428-5c3a-5f7a687d9f7c \
    | openssl x509 -noout -text
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            3e:bf:14:4d:1c:f1:96:b8:bf:6b:72:be:3b:77:e9:c0:2d:06:84:44
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: CN = example.com
        Validity
            Not Before: Nov 20 21:10:11 2023 GMT
            Not After : Nov 21 21:30:41 2023 GMT
        Subject: CN = example.com
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                Public-Key: (2048 bit)
                Modulus:
                    00:c3:2a:b8:0d:34:76:46:ca:44:9c:1c:cb:74:7e:
                    41:b5:ee:55:2f:db:e7:d4:b7:19:33:24:42:c5:23:
                    c1:ce:48:20:9e:f4:04:94:03:01:d4:2e:d3:1e:13:
                    53:99:3a:6a:77:55:df:74:7c:52:02:0c:54:d4:f3:
                    0d:e3:dd:75:d9:4d:2c:98:3d:5b:c1:03:72:f2:30:
                    3e:0f:d9:51:17:92:0b:0f:9f:be:6f:82:f3:07:0a:
                    a9:4b:77:3f:12:7a:cf:4e:70:ec:92:81:e7:8f:ec:
                    0f:2c:02:9f:05:f8:2d:18:38:5c:31:c4:16:7b:ce:
                    96:f5:46:99:3a:8e:8b:07:4e:4d:e9:79:86:34:bc:
                    7c:34:6f:17:c3:38:1c:76:44:aa:05:22:8b:5f:98:
                    06:19:bd:0e:53:e9:1e:89:46:15:25:72:67:68:80:
                    ff:46:01:6e:04:b9:49:bf:bf:80:a5:33:5c:89:8a:
                    3a:33:62:b6:8f:d7:f4:ef:93:09:1f:47:ab:1a:22:
                    30:6e:c7:8c:2e:2b:dc:b3:32:f8:06:7c:2e:eb:d5:
                    26:9a:7b:9a:8a:f5:15:18:81:d2:26:74:cf:b6:5b:
                    11:8a:f8:90:49:93:26:9e:f9:09:63:36:bb:e1:ca:
                    1a:54:37:bb:fb:a0:4a:77:4f:87:07:86:26:71:b3:
                    66:2b
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Key Usage: critical
                Certificate Sign, CRL Sign
            X509v3 Basic Constraints: critical
                CA:TRUE
            X509v3 Subject Key Identifier:
                AE:5A:43:38:E7:BC:57:02:74:E9:7F:54:E6:9C:69:9A:4A:B3:0F:4F
            X509v3 Authority Key Identifier:
                AE:5A:43:38:E7:BC:57:02:74:E9:7F:54:E6:9C:69:9A:4A:B3:0F:4F
            X509v3 Subject Alternative Name:
                DNS:example.com
    Signature Algorithm: sha256WithRSAEncryption
    Signature Value:
        b9:a3:b2:05:e8:5b:17:8f:81:00:e0:a2:35:96:1a:fa:2a:fe:
        b4:90:98:5a:23:94:9a:08:8c:5c:f6:8f:0a:4d:28:41:90:64:
        50:0f:4d:85:ba:4f:73:c5:d2:65:24:29:f2:b6:d2:03:bf:32:
        e4:a5:6f:df:41:88:45:3e:08:f0:22:2a:07:d5:ab:d5:45:f7:
        79:1e:1d:ef:90:b0:f1:b0:b3:0d:fb:d9:72:cb:ab:92:f4:ac:
        43:ad:6c:ff:71:d6:54:19:f3:0e:a0:b9:92:e9:30:22:76:1c:
        b8:67:4e:b0:96:4f:48:28:c4:db:3b:73:be:2c:12:2b:ee:82:
        29:c4:7d:a5:68:e0:08:c1:d4:c2:96:74:70:15:90:b0:1a:49:
        0f:b2:9d:3c:97:79:f9:64:a8:3d:a2:1c:bf:fc:7d:69:5f:1d:
        20:ae:7e:c6:5e:19:4e:34:f5:14:8c:ff:2b:0d:97:dc:fd:9e:
        44:a3:92:a5:05:10:88:71:6e:16:71:db:0f:74:7f:47:17:14:
        2f:3b:da:fd:33:44:bd:88:fc:b3:95:b6:40:76:df:91:86:f3:
        3c:98:d5:3e:08:6a:5a:d4:ec:44:e6:2d:98:d1:e0:08:4d:fb:
        7c:8d:fa:00:f8:c3:32:0c:2d:0d:98:5d:89:5b:f2:44:44:fc:
        05:41:d9:8a
```

После чего для данного сертификата можно определить роль, с помощью которой
в дальнейшем можно будет выписывать сертификаты:
```console
$ vault write pki/roles/2023-servers allow_any_name=true
Key                                   Value
---                                   -----
allow_any_name                        true
allow_bare_domains                    false
allow_glob_domains                    false
allow_ip_sans                         true
allow_localhost                       true
allow_subdomains                      false
allow_token_displayname               false
allow_wildcard_certificates           true
allowed_domains                       []
allowed_domains_template              false
allowed_other_sans                    []
allowed_serial_numbers                []
allowed_uri_sans                      []
allowed_uri_sans_template             false
allowed_user_ids                      []
basic_constraints_valid_for_non_ca    false
client_flag                           true
cn_validations                        [email hostname]
code_signing_flag                     false
country                               []
email_protection_flag                 false
enforce_hostnames                     true
ext_key_usage                         []
ext_key_usage_oids                    []
generate_lease                        false
issuer_ref                            default
key_bits                              2048
key_type                              rsa
key_usage                             [DigitalSignature KeyAgreement KeyEncipherment]
locality                              []
max_ttl                               0s
no_store                              false
not_after                             n/a
not_before_duration                   30s
organization                          []
ou                                    []
policy_identifiers                    []
postal_code                           []
province                              []
require_cn                            true
server_flag                           true
signature_bits                        256
street_address                        []
ttl                                   0s
use_csr_common_name                   true
use_csr_sans                          true
use_pss                               false
```

Теперь можно выписывать сертификаты следующей командой:
```console
$ vault write pki/issue/2023-servers common_name="test.example.com" ttl="24h"
Key                 Value
---                 -----
ca_chain            [-----BEGIN CERTIFICATE-----
MIIDNTCCAh2gAwIBAgIUPr8UTRzxlri/a3K+O3fpwC0GhEQwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMjMxMTIwMjExMDExWhcNMjMx
MTIxMjEzMDQxWjAWMRQwEgYDVQQDEwtleGFtcGxlLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAMMquA00dkbKRJwcy3R+QbXuVS/b59S3GTMkQsUj
wc5IIJ70BJQDAdQu0x4TU5k6andV33R8UgIMVNTzDePdddlNLJg9W8EDcvIwPg/Z
UReSCw+fvm+C8wcKqUt3PxJ6z05w7JKB54/sDywCnwX4LRg4XDHEFnvOlvVGmTqO
iwdOTel5hjS8fDRvF8M4HHZEqgUii1+YBhm9DlPpHolGFSVyZ2iA/0YBbgS5Sb+/
gKUzXImKOjNito/X9O+TCR9HqxoiMG7HjC4r3LMy+AZ8LuvVJpp7mor1FRiB0iZ0
z7ZbEYr4kEmTJp75CWM2u+HKGlQ3u/ugSndPhweGJnGzZisCAwEAAaN7MHkwDgYD
VR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFK5aQzjnvFcC
dOl/VOacaZpKsw9PMB8GA1UdIwQYMBaAFK5aQzjnvFcCdOl/VOacaZpKsw9PMBYG
A1UdEQQPMA2CC2V4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQC5o7IF6FsX
j4EA4KI1lhr6Kv60kJhaI5SaCIxc9o8KTShBkGRQD02Fuk9zxdJlJCnyttIDvzLk
pW/fQYhFPgjwIioH1avVRfd5Hh3vkLDxsLMN+9lyy6uS9KxDrWz/cdZUGfMOoLmS
6TAidhy4Z06wlk9IKMTbO3O+LBIr7oIpxH2laOAIwdTClnRwFZCwGkkPsp08l3n5
ZKg9ohy//H1pXx0grn7GXhlONPUUjP8rDZfc/Z5Eo5KlBRCIcW4WcdsPdH9HFxQv
O9r9M0S9iPyzlbZAdt+RhvM8mNU+CGpa1OxE5i2Y0eAITft8jfoA+MMyDC0NmF2J
W/JERPwFQdmK
-----END CERTIFICATE-----]
certificate         -----BEGIN CERTIFICATE-----
MIIDTzCCAjegAwIBAgIUHTTWLpiBAl3/xIvzwZvbOfar30QwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMjMxMTIwMjExOTE4WhcNMjMx
MTIxMjExOTQ4WjAbMRkwFwYDVQQDExB0ZXN0LmV4YW1wbGUuY29tMIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4lec07w+zRA14s11dYJA8be4bw8qKUdr
w2Pz1UqdCwKOSsdyRyx/FJn1QsdHiqZfQEFug8v9xH0PJSW9VWyzT7drZmyU9qc0
5ajTatQEPjSfEQRuoMV75ni7xDH7YgEfk+m6fjOyrHusXM0H6tv8HmbherIpnO+O
xnjqXdEDDioSsaaiWv70NsIEAr6z+P7q6z+I85NMyaODcRyPf4bFqQaKRqbDbABu
h6tpkhG6Zn/09sCPoCa/HroJBpuRfC6oay4Q9bj8DT0xsq+6b3TH2fK8fTpExu0c
r9nEv9Nbb/Qv0y8Rezyv8gCl09OLFaE/wcUN7qdREfu+8JTx6lJV1wIDAQABo4GP
MIGMMA4GA1UdDwEB/wQEAwIDqDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUH
AwIwHQYDVR0OBBYEFPAMiaQiLgckjrJNzA5RV091yulKMB8GA1UdIwQYMBaAFK5a
QzjnvFcCdOl/VOacaZpKsw9PMBsGA1UdEQQUMBKCEHRlc3QuZXhhbXBsZS5jb20w
DQYJKoZIhvcNAQELBQADggEBAERd7d7RfqIPQKZZ76mhUW098g9cRoMU7TOwrLdl
VfMFfIxme7KFCWhJ1FAinZkpFdBikQG00sm3myyvivxTvE9cQPZGkV3hAqYI+D+g
zJBupzr29kbLKBmDlmG/U1KnhNCHI1TptQ8ifRoln5+BY2MnB1BfArh0RE33Lw3Q
5M11x0GvKJ4cpSDkvz4+MtRO9OBwCDIAks5bnaDrGSlCee8U8pWRTch6amYwHBuH
1jhovGOe1Qf7AUJEWXs/07QmaXagqv8r+vckEtrAp2odUXcBazKdyxByHUiSv0a8
CIBluKGaA/B2UciNPX4lMCV2Njq04/HLiZc73bVKAho5B9Q=
-----END CERTIFICATE-----
expiration          1700601588
issuing_ca          -----BEGIN CERTIFICATE-----
MIIDNTCCAh2gAwIBAgIUPr8UTRzxlri/a3K+O3fpwC0GhEQwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMjMxMTIwMjExMDExWhcNMjMx
MTIxMjEzMDQxWjAWMRQwEgYDVQQDEwtleGFtcGxlLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAMMquA00dkbKRJwcy3R+QbXuVS/b59S3GTMkQsUj
wc5IIJ70BJQDAdQu0x4TU5k6andV33R8UgIMVNTzDePdddlNLJg9W8EDcvIwPg/Z
UReSCw+fvm+C8wcKqUt3PxJ6z05w7JKB54/sDywCnwX4LRg4XDHEFnvOlvVGmTqO
iwdOTel5hjS8fDRvF8M4HHZEqgUii1+YBhm9DlPpHolGFSVyZ2iA/0YBbgS5Sb+/
gKUzXImKOjNito/X9O+TCR9HqxoiMG7HjC4r3LMy+AZ8LuvVJpp7mor1FRiB0iZ0
z7ZbEYr4kEmTJp75CWM2u+HKGlQ3u/ugSndPhweGJnGzZisCAwEAAaN7MHkwDgYD
VR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFK5aQzjnvFcC
dOl/VOacaZpKsw9PMB8GA1UdIwQYMBaAFK5aQzjnvFcCdOl/VOacaZpKsw9PMBYG
A1UdEQQPMA2CC2V4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQC5o7IF6FsX
j4EA4KI1lhr6Kv60kJhaI5SaCIxc9o8KTShBkGRQD02Fuk9zxdJlJCnyttIDvzLk
pW/fQYhFPgjwIioH1avVRfd5Hh3vkLDxsLMN+9lyy6uS9KxDrWz/cdZUGfMOoLmS
6TAidhy4Z06wlk9IKMTbO3O+LBIr7oIpxH2laOAIwdTClnRwFZCwGkkPsp08l3n5
ZKg9ohy//H1pXx0grn7GXhlONPUUjP8rDZfc/Z5Eo5KlBRCIcW4WcdsPdH9HFxQv
O9r9M0S9iPyzlbZAdt+RhvM8mNU+CGpa1OxE5i2Y0eAITft8jfoA+MMyDC0NmF2J
W/JERPwFQdmK
-----END CERTIFICATE-----
private_key         -----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA4lec07w+zRA14s11dYJA8be4bw8qKUdrw2Pz1UqdCwKOSsdy
Ryx/FJn1QsdHiqZfQEFug8v9xH0PJSW9VWyzT7drZmyU9qc05ajTatQEPjSfEQRu
oMV75ni7xDH7YgEfk+m6fjOyrHusXM0H6tv8HmbherIpnO+OxnjqXdEDDioSsaai
Wv70NsIEAr6z+P7q6z+I85NMyaODcRyPf4bFqQaKRqbDbABuh6tpkhG6Zn/09sCP
oCa/HroJBpuRfC6oay4Q9bj8DT0xsq+6b3TH2fK8fTpExu0cr9nEv9Nbb/Qv0y8R
ezyv8gCl09OLFaE/wcUN7qdREfu+8JTx6lJV1wIDAQABAoIBAHm7IxZhZOlG8aRE
WgrT/ffClofsgQFobxXL+QTatzGHs12WfOi4jrRWxSigIuL37bySNEzf9mCp3e5d
qMF7z8hs3m9csJUjgniN3v7SfmGyTWaKmrEO5T8j6eBv/UJNVl1n1Cwxw8OuVBop
DzwKCkOTk2s4zNahpIkr2OiSE/GWPiiXn5rHcV6MdMvbI5qc/RXp/bioo7X+0kg6
BPdSQbJP2K9CTJzjh4mGqkJI6hOpJ9ChepyU8CUpyogi3THwcM0c5ru2/U+52xjO
kjfRffHRbwOgkUWbIjXVCi+6C5esrJqZVrhbrvwe58TGTa5QBoQATouvyiAzl+/U
qhHv2MkCgYEA6lDmQkVvZQd5udghNrZ+nFyxKfJdKm2lp3TP1HshI2cxMm3SxVga
qDGvJGSgjaKgCGxAeYQVVISVILQ1frPQ7aRn3BGJPvgUY3UGk8pL4ZMhAwZ7eNUD
R2bWZc3QxN2yH5YBQvOSTDpMVayqjHAVoK7wo/EyYfWAF5GfRsBU8cMCgYEA90nP
I79P3Ik4xB//QseXh2mdpd2KRDcIkf1MwhidnB9/QSY1YiV5GipVSZgC4hUeBPmz
05C5CrcCBoASstqgxskWB0Zp6HPH6wixok17ijy+H23YB3bYf2STFZVLvTSVt4xT
F8Po08mdo4NZaCFjsqAcigjbd7Mp83K18YnVVl0CgYAvfbQdrHsWa/x0+WRJ9ZUV
1gema9QMPGr91MQm2cnupgSnpvC4RNIqUt+frbGI43QyINa0ilvUZIbhOQU6p/Fp
qQ/P39IEbD2dpuNtYuwcTTi8pzyxUeM3PpWnzp5IuHJYyot46Ws2ff5owvVSP4ly
puJpKALBLgQHQuGYcnUFBwKBgF/ZMfqPGqtGXMRYCp6dsjQAUeSKXB9YnW/ImEnb
NKvg4XglESf7klb79ZbS3rs2qC4RgwwL2k025ggS+Cxu5UZnhqxHNKGuztxgwElC
cxH/vUl9T/CEtiGaoBALkBHEIgvEzig1/TapvPo13R+pYXVI7gbqq/ZXcXk1CySV
4iTZAoGBAIbhooHtQbiAS4BF9ffQAoMxsbTGZeMUAu5snvI2IFo/gb5vgO/XF91e
+6e/z9+R7knvX331UKniHFtPvBwwIB7oSRllEInY/t++TS+eZZrYHl9kPX3F/xbw
/NzljHESdpucYFdALpI9JZZZ1ustAxd8jIXHRg++PMDAp9RefFAs
-----END RSA PRIVATE KEY-----
private_key_type    rsa
serial_number       1d:34:d6:2e:98:81:02:5d:ff:c4:8b:f3:c1:9b:db:39:f6:ab:df:44
```

Посмотреть сертификаты в движке можно следующим образом:
```console
$ vault list pki/certs
Keys
----
1d:34:d6:2e:98:81:02:5d:ff:c4:8b:f3:c1:9b:db:39:f6:ab:df:44
3e:bf:14:4d:1c:f1:96:b8:bf:6b:72:be:3b:77:e9:c0:2d:06:84:44
$ vault read -field=certificate \
    pki/cert/1d:34:d6:2e:98:81:02:5d:ff:c4:8b:f3:c1:9b:db:39:f6:ab:df:44 \
    | openssl x509 -noout -text
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            1d:34:d6:2e:98:81:02:5d:ff:c4:8b:f3:c1:9b:db:39:f6:ab:df:44
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: CN = example.com
        Validity
            Not Before: Nov 20 21:19:18 2023 GMT
            Not After : Nov 21 21:19:48 2023 GMT
        Subject: CN = test.example.com
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                Public-Key: (2048 bit)
                Modulus:
                    00:e2:57:9c:d3:bc:3e:cd:10:35:e2:cd:75:75:82:
                    40:f1:b7:b8:6f:0f:2a:29:47:6b:c3:63:f3:d5:4a:
                    9d:0b:02:8e:4a:c7:72:47:2c:7f:14:99:f5:42:c7:
                    47:8a:a6:5f:40:41:6e:83:cb:fd:c4:7d:0f:25:25:
                    bd:55:6c:b3:4f:b7:6b:66:6c:94:f6:a7:34:e5:a8:
                    d3:6a:d4:04:3e:34:9f:11:04:6e:a0:c5:7b:e6:78:
                    bb:c4:31:fb:62:01:1f:93:e9:ba:7e:33:b2:ac:7b:
                    ac:5c:cd:07:ea:db:fc:1e:66:e1:7a:b2:29:9c:ef:
                    8e:c6:78:ea:5d:d1:03:0e:2a:12:b1:a6:a2:5a:fe:
                    f4:36:c2:04:02:be:b3:f8:fe:ea:eb:3f:88:f3:93:
                    4c:c9:a3:83:71:1c:8f:7f:86:c5:a9:06:8a:46:a6:
                    c3:6c:00:6e:87:ab:69:92:11:ba:66:7f:f4:f6:c0:
                    8f:a0:26:bf:1e:ba:09:06:9b:91:7c:2e:a8:6b:2e:
                    10:f5:b8:fc:0d:3d:31:b2:af:ba:6f:74:c7:d9:f2:
                    bc:7d:3a:44:c6:ed:1c:af:d9:c4:bf:d3:5b:6f:f4:
                    2f:d3:2f:11:7b:3c:af:f2:00:a5:d3:d3:8b:15:a1:
                    3f:c1:c5:0d:ee:a7:51:11:fb:be:f0:94:f1:ea:52:
                    55:d7
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Key Usage: critical
                Digital Signature, Key Encipherment, Key Agreement
            X509v3 Extended Key Usage:
                TLS Web Server Authentication, TLS Web Client Authentication
            X509v3 Subject Key Identifier:
                F0:0C:89:A4:22:2E:07:24:8E:B2:4D:CC:0E:51:57:4F:75:CA:E9:4A
            X509v3 Authority Key Identifier:
                AE:5A:43:38:E7:BC:57:02:74:E9:7F:54:E6:9C:69:9A:4A:B3:0F:4F
            X509v3 Subject Alternative Name:
                DNS:test.example.com
    Signature Algorithm: sha256WithRSAEncryption
    Signature Value:
        44:5d:ed:de:d1:7e:a2:0f:40:a6:59:ef:a9:a1:51:6d:3d:f2:
        0f:5c:46:83:14:ed:33:b0:ac:b7:65:55:f3:05:7c:8c:66:7b:
        b2:85:09:68:49:d4:50:22:9d:99:29:15:d0:62:91:01:b4:d2:
        c9:b7:9b:2c:af:8a:fc:53:bc:4f:5c:40:f6:46:91:5d:e1:02:
        a6:08:f8:3f:a0:cc:90:6e:a7:3a:f6:f6:46:cb:28:19:83:96:
        61:bf:53:52:a7:84:d0:87:23:54:e9:b5:0f:22:7d:1a:25:9f:
        9f:81:63:63:27:07:50:5f:02:b8:74:44:4d:f7:2f:0d:d0:e4:
        cd:75:c7:41:af:28:9e:1c:a5:20:e4:bf:3e:3e:32:d4:4e:f4:
        e0:70:08:32:00:92:ce:5b:9d:a0:eb:19:29:42:79:ef:14:f2:
        95:91:4d:c8:7a:6a:66:30:1c:1b:87:d6:38:68:bc:63:9e:d5:
        07:fb:01:42:44:59:7b:3f:d3:b4:26:69:76:a0:aa:ff:2b:fa:
        f7:24:12:da:c0:a7:6a:1d:51:77:01:6b:32:9d:cb:10:72:1d:
        48:92:bf:46:bc:08:80:65:b8:a1:9a:03:f0:76:51:c8:8d:3d:
        7e:25:30:25:76:36:3a:b4:e3:f1:cb:89:97:3b:dd:b5:4a:02:
        1a:39:07:d4
```

Таким образом можно построить инфраструктуру управления сертификатами, например,
внутри компании.

## Agent
Утилита `vault` может также работать в виде агента на сервере, синхронизируя
секреты, а также вставляя их в конфигурационные файлы. Создадим конфигурацию
агента в файле `config.hcl`:
```
auto_auth {
  method {
     type = "token_file"
     config = {
        token_file_path = "/home/vagrant/.vault-token"
     }
  }
  sink "file" {
    config = {
      path = "/tmp/file-foo"
    }
  }
}

template {
  contents = <<EOF
{{- with secret "secret/test" -}}
secret_value={{ .Data.data.somekey }}
{{- end }}
EOF
  destination = "/home/vagrant/config"
}
```
Здесь мы указали параметры авторизации, а также шаблон конфигурации, который
использует key-value движок `secret/` и путь, по которому полученная конфигурация
будет сохранена. Запустим агент, добавив в него опцию `-exit-after-auth`, чтобы
после создания конфигурации агент завершил свою работу, а не был постоянно запущен
отслеживая секреты:
```console
$ vault agent -exit-after-auth -config config.hcl
==> Vault Agent started! Log data will stream in below:

==> Vault Agent configuration:

           Api Address 1: http://bufconn
                     Cgo: disabled
               Log Level:
                 Version: Vault v1.13.3, built 2023-06-06T18:12:37Z
             Version Sha: 3bedf816cbf851656ae9e6bd65dd4a67a9ddff5e

2023-11-20T22:09:51.206Z [INFO]  agent.sink.file: creating file sink
2023-11-20T22:09:51.207Z [INFO]  agent.sink.file: file sink configured: path=/tmp/file-foo mode=-rw-r-----
2023-11-20T22:09:51.207Z [INFO]  agent.template.server: starting template server
2023-11-20T22:09:51.207Z [INFO] (runner) creating new runner (dry: false, once: false)
2023-11-20T22:09:51.207Z [INFO]  agent.auth.handler: starting auth handler
2023-11-20T22:09:51.207Z [INFO]  agent.auth.handler: authenticating
2023-11-20T22:09:51.208Z [INFO]  agent.sink.server: starting sink server
2023-11-20T22:09:51.208Z [INFO] (runner) creating watcher
2023-11-20T22:09:51.209Z [INFO]  agent.auth.handler: authentication successful, sending token to sinks
2023-11-20T22:09:51.209Z [INFO]  agent.auth.handler: not starting token renewal process, as token has unlimited TTL
2023-11-20T22:09:51.209Z [INFO]  agent.template.server: template server received new token
2023-11-20T22:09:51.209Z [INFO] (runner) stopping
2023-11-20T22:09:51.209Z [INFO] (runner) creating new runner (dry: false, once: false)
2023-11-20T22:09:51.209Z [INFO]  agent.sink.file: token written: path=/tmp/file-foo
2023-11-20T22:09:51.210Z [INFO]  agent.sink.server: sink server stopped
2023-11-20T22:09:51.210Z [INFO]  agent: sinks finished, exiting
2023-11-20T22:09:51.210Z [INFO] (runner) creating watcher
2023-11-20T22:09:51.210Z [INFO] (runner) starting
2023-11-20T22:09:51.216Z [INFO] (runner) stopping
2023-11-20T22:09:51.216Z [INFO]  agent.template.server: template server stopped
2023-11-20T22:09:51.216Z [INFO] (runner) received finish
2023-11-20T22:09:51.216Z [INFO]  agent.auth.handler: shutdown triggered, stopping lifetime watcher
2023-11-20T22:09:51.216Z [INFO]  agent.auth.handler: auth handler stopped

$ cat /home/vagrant/config
secret_value=anothervalue
```

Как видно агент взял информацию из движка `secret/` и вставил в описанный шаблон.
Таким образом можно шаблонизировать конфигурации различных приложений, передавая
в них чувствительные данные.

## Transit
Vault сервер также предоставляет возможность использования движка
[transit][transit-engine], который позволяет шифровать данные отправляя их
на сервер и возвращая их в зашифрованном виде, при этом сами данные не сохраняются
на сервер. Это можно использовать с такими приложениями как [sops][], управляя
доступом к движку для определенного круга лиц на стороне vault сервера.

Активируем движок и сгенерируем в нем ключ:
```console
$ vault secrets enable transit
Success! Enabled the transit secrets engine at: transit/
$ vault write transit/keys/testkey type=rsa-4096
Success! Data written to: transit/keys/testkey
```

После чего воспользуемся командой `sops`. Ей можно явно задать в опциях путь
до движка опцией `--hc-vault-transit`:
```console
$ echo secret | sops --hc-vault-transit $VAULT_ADDR/v1/transit/keys/testkey \
    -e /dev/stdin
{
        "data": "ENC[AES256_GCM,data:0Fllhj06Vw==,iv:2H4+9YFWwIYaLNluDsUR28L44t6R3t3UePw5v4mK/xA=,tag:woPdlSGAPERYB9ZJecR5TQ==,type:str]",
        "sops": {
                "kms": null,
                "gcp_kms": null,
                "azure_kv": null,
                "hc_vault": [
                        {
                                "vault_address": "http://0.0.0.0:8200",
                                "engine_path": "transit",
                                "key_name": "testkey",
                                "created_at": "2023-11-21T20:19:41Z",
                                "enc": "vault:v1:mcs+8rnEdOABX+gzA0hzVCDqM3ELJbuG6JNQwTY/FkBMrc8pI7z34Kn5kzrHuW1Wj7+s3cEPt6rEfN4XCFw7cZyzD8KA2sHtXxB/ygIimW4hZoBHJC0TyKzK3792z8uVW9gsPx14asHVFxsarxVW/2psXRG7JT78rNJisNtkZbf4zuoFh2I0Hkv6oYYSp/q3slJ/tOcE408tLT9d9hu3PfJmwFA+wT+YQx2jpx8v1nGVok4HsdOXCwV13zvqsF1LlGIINIqHSFKBNqdCzpEjdFNma642A5K8GiDaJ6mNsX8BLBuiZ/9QETLdMEKOE60BEt9OE0uH03jnPJM/dy1spuan2f3UIQXuQtMHvfSDoELGcwf8MzaMj/0If9NAoJvWyPkdotuN0LEbYKFAN0xCgnlznn6dNhNRqocU7VBrorgNg80C5haNOma5HcudOSSmCT6zvWMVDeiICbqNp69WgEMxQ6L5Sk5m9Fe9mFbAjcSApqN/XeM1a5Lc63umu4zJrfIMqZhta1QfdnVC4r5W7UlBZFCbaoaBaTTHXDiQBSbuzJDfUTLk4VCa8PhTJWlczfH+tYTeN1075MtquEZ4nuo1RCPrfpU4H3i9dlOIm7UCMG5uQTRTLenlIiD8UDF6tRfGqrcDMoVe3ZpWuhWMFICuJCHBbW4FYpXm30H63hA="
                        }
                ],
                "age": null,
                "lastmodified": "2023-11-21T20:19:41Z",
                "mac": "ENC[AES256_GCM,data:grPZ9LLSKxdjbIIbbbx8ElkJj7eL408YJRFoezBnGLR4Wc2WCx8Yw8Qi8Fcbu1PoHrHwjilak4TPY2Pw3vtPlsjOOaFUx6b8VL37wH+CAygxjXiu0YiBgvD0TQwqSq/y9A9ultcCteb++cXywafG9weyUfooUIA8f7Fu7PA7ofc=,iv:E3ekXNpfXVc1e5kjMBVrjEYnhzv70Qd2vehDu61E46Y=,tag:LQSRYDROrH0Foi3KyOb/zw==,type:str]",
                "pgp": null,
                "unencrypted_suffix": "_unencrypted",
                "version": "3.8.1"
        }
}
```

Также можно задать через файл конфигурации `.sops.yaml`:
```console
$ cat <<EOF> .sops.yaml
> creation_rules:
> - hc_vault_transit_uri: $VAULT_ADDR/v1/transit/keys/testkey
> EOF
$ sops -e -i config
$ cat config
{
        "data": "ENC[AES256_GCM,data:fSVS3USuD4hLtfbVOMI5iNqY7o5p0/M1p98=,iv:jxwNyzWhoOnh11TvtwBMEOXm/QncMtZhLh2mT0YVPMg=,tag:i/73k3yh3yy0guiwlHiSKA==,type:str]",
        "sops": {
                "kms": null,
                "gcp_kms": null,
                "azure_kv": null,
                "hc_vault": [
                        {
                                "vault_address": "http://0.0.0.0:8200",
                                "engine_path": "transit",
                                "key_name": "testkey",
                                "created_at": "2023-11-21T20:35:10Z",
                                "enc": "vault:v1:C5DsMajw83NQYuM9t6+4Ohgd/WsW76Caimvjjh+FxFa5mZjKTdVur2YLTcLZdtW9IhNa5VJM+ZpaDlJ3aog8Rgmp49Id3eVqYB9ZpTBwE7yIAaUR8Bx8OpqMZk7WWr+cG6RVRW51yoSzpx8qMY1HKrHTn+IABVenPGN0ALCbis28iSHTDrazv1N1nUgDzPyqNOKww9f+koIeXylG9z/rrTTL/coLBSoRjQ3h2iekOz3GeyVVdup9v8LeGr0MXVNun463JYDY+G8hM3Bn5HBg0R0SuKrZDoVV5O0+7w/MtdwVtQfWZeZ0bZY6mRwozgA53xt5EP7eAYD5dvK4jka0TrXGzDBe/kNuoASww4H7NrMKSdZvw9a6FiXgtL5Ki9BX6jWW+L9XIH1hP2eI6ol5Jfe2Z3k78Bacmkxf8bwcZi+2HGDJkW/cUbvoe8zKqhPN1BkXXFHsGVfeSaflrrNAnmfSJomqPpzwaPqlppv8NBQHoicNAnIIxszzuU6nvFBt4mlxfiYuJLPiyYQpvfdpohOWifT5NqRQ3fjdcx23ytFlZ8RjJDAbc9fHwqlfcug/Cp3E1gWqqoi1PRqfUWkgEz7Rt3y+JobhcD8+HTOYfO2nCxlEHL0L/hwqJ3QnPSpLjfPfDL001JRwNDQWrT+cLcpDx9Ju2tGuc66o5/42rEw="
                        }
                ],
                "age": null,
                "lastmodified": "2023-11-21T20:35:10Z",
                "mac": "ENC[AES256_GCM,data:D6gFwf4VxO+EsaIMRpAwxQaZRtRnB+DYziUDVkrY6Kdtndk/wlTDgbqxCmJlJCOmSp3phR6aFEUUtxPs9U+vHjo3bX/6viecscmDOSDj26xKpNyu08yv3KjH97vabrMRIgSYR6/1n8w9NTFS4RxdUkYDX8zauHRqx3Btz9FRmik=,iv:fXuXuFWU/FRW4o/Ao98Z3/5/9+ayiGy+0ZALp1UHI1c=,tag:z/Y3Jw9FATexgzaVNDdkgQ==,type:str]",
                "pgp": null,
                "unencrypted_suffix": "_unencrypted",
                "version": "3.8.1"
        }
}
$ sops -d -i config
$ cat config
secret_value=anothervalue
```

[vault]:https://developer.hashicorp.com/vault/docs/what-is-vault
[kv-engine]:https://developer.hashicorp.com/vault/tutorials/secrets-management/static-secrets
[pki-engine]:https://developer.hashicorp.com/vault/tutorials/secrets-management/pki-engine
[transit-engine]:https://developer.hashicorp.com/vault/tutorials/encryption-as-a-service/eaas-transit
[sops]:https://github.com/getsops/sops
