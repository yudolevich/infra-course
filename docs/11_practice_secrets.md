# Secret Management
В данном практическом занятии познакомимся с инструментами для локального
управления чувствительными данными(пароли, ключи, токены и т.д.).

## Vagrant
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "node" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "node"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns ansible
      curl -L https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-linux-amd64.tar.gz \
        | tar xvz --strip-components=1 -C /usr/local/bin age/age age/age-keygen
      curl -L https://github.com/getsops/sops/releases/download/v3.8.1/sops-v3.8.1.linux.amd64 \
        -o /usr/local/bin/sops && chmod +x /usr/local/bin/sops
    SHELL
  end
end
```

## Ansible Vault
Различные стеки технологий часто имеют собственные инструменты для управления
секретами, в [ansible][] таким инструментом является [ansible-vault][]. Утилита
`ansible-vault` позволяет шифровать данные парольной фразой, так что при утечке
не будет возможности воспользоваться секретными данными.

### Create
Для создания секрета можно воспользоваться подкомандой `create`, которая запросит
пароль для нового файла и откроет текстовый редактор, определенный в переменной
`EDITOR`:
```console
$ EDITOR=nano ansible-vault create secret.yaml
New Vault password:
Confirm New Vault password:
  /home/vagrant/.ansible/tmp/ansible-local-71380mr2u0wp/tmpfx3ulqj0.yaml *
item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value
  key2: secret_value

^G Help      ^O Write Out ^W Where Is  ^K Cut       ^T Execute   ^C Location
^X Exit      ^R Read File ^\ Replace   ^U Paste     ^J Justify   ^/ Go To Line
```
Внесем данные, которые хотим зашифровать и сохраним файл.

Если на текущий момент посмотреть данный файл, то содержимое будет в зашифрованном
виде:
```console
$ cat secret.yaml
$ANSIBLE_VAULT;1.1;AES256
31613863313562623337353430663066646236643639626635356135653165656130343036656130
6638653365303364623965366461626262373064353462610a356164393364363266343264633938
65323664353434616438336163636165393434633333643433636435636262353338663839626330
3138663266346232650a373661333662366166363739303032373430373731316237386232653133
65303366363064373039643730653339396439353736376635626434653536633361666665643938
63373231383166363136303566316438323438333837323366623161346332376562346130336630
35316461623263663236313961383636663936336334616132663066343562616534646163373837
35616265313031346638653834666533393164636334643438363634303565656138373030396666
33363035323065386166353663653462653335383232376634643064343564343338383837326666
3932306430363162613064346437303630633239656136353831
```

### View
Для вывода на экран расшифрованное содержимое можно воспользоваться подкомандой
`view`:
```console
$ ansible-vault view secret.yaml
Vault password:
item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value
  key2: secret_value
```

### Edit
Отредактировать же можно подкомандой `edit`, которая также откроет текстовый
редактор как в команде `create`:
```console
$ ansible-vault edit secret.yaml
Vault password:

item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value1
  key2: secret_value2
~
~
<ocal-71616c9rq5j8/tmpjkj4ar9i.yaml" 8L, 119B written   8,21  All
```

### Encrypt/Decrypt
Расшифровать и зашифровать файл на месте можно соответственно подкомандами
`decrypt` и `encrypt`:
```console
$ cat secret.yaml
$ANSIBLE_VAULT;1.1;AES256
34366264616236613863303231396132373635323130323065396661633236363866663634393236
3363316662656535633930636632613733373932353935350a363766353231373364346230646634
31363936663633373630326131356435323037663533616265326234643264353631316166373761
3932386630653432620a623063383932623436326663633435393232316166393130613465393231
39346339333839653232323431346539616362326139356636356639316630666435376233616431
32666438356466346633633835363232383534316164343636633136313863643163666663383766
30313130376161386366636565653334313036313332313435356330636234633732616366653263
63393035383765623562626531633238306133633066666434386666656131666663623062356366
37623233623162313339333237356533326566663038323561336138393462386633616538323436
3633663062373631666539653036633034383834376161366161
$ ansible-vault decrypt secret.yaml
Vault password:
Decryption successful
$ cat secret.yaml
item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value1
  key2: secret_value2
$ ansible-vault encrypt secret.yaml
New Vault password:
Confirm New Vault password:
Encryption successful
$ cat secret.yaml
$ANSIBLE_VAULT;1.1;AES256
64323737633536313039393432653135396563333538623538376133383062323961353033346330
3439633434393039303633393066373437643938626130320a326466323566306535373231636639
65653833613634313835646339303935653336663037306238386334373661393830633736383866
3039363536303833640a373735636334373535383034666533633533316539646165333034343739
66383335613736316565356535383137373562386664656232636332616261623132353834356433
31626332623033613535383264646534393436356230386262323135353738343634393939656531
38663339383330656637623037663930386539383639336337646230373534623165656665376562
64366463343939613863643333313538623365366632393131323137623535666164343565376361
30343538356464323461643734363432643935653263336631343062333835396161313135343164
3734643933313330396133643230646639383439386566396436
```

### Playbook
Зашифрованные файлы с помощью `ansible-vault` можно использовать в плейбуках.
Создадим файл `playbook.yaml`, который будет использовать наш зашифрованный файл:
```yaml
---
- hosts: localhost
  connection: local
  gather_facts: False
  vars_files:
  - secret.yaml
  tasks:
  - name: test
    debug:
      msg: "{{ list }}"
```
И запустим его указав опцию `--ask-vault-pass`:
```console
$ ansible-playbook playbook.yaml --ask-vault-pass
Vault password:

PLAY [localhost] *****************************************************************

TASK [test] **********************************************************************
ok: [localhost] => {
    "msg": [
        "secret_item1",
        "secret_item2",
        "secret_item3"
    ]
}

PLAY RECAP ***********************************************************************
localhost                  : ok=1    changed=0    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
```
Таким образом во время запуска мы можем получить доступ к секретным данным и
использовать их в плейбуках.

### Encrypt string
Также с помощью подкоманды `encrypt_string` можно зашифровать переданную строку,
которую можно указать как аргумент или же через стандартный ввод: 
```console
$ ansible-vault encrypt_string
New Vault password:
Confirm New Vault password:
Reading plaintext input from stdin. (ctrl-d to end input, twice if your content does not already have a newline)
secretString
Encryption successful
!vault |
          $ANSIBLE_VAULT;1.1;AES256
          65623232363334306166353539343338613238646362643462356237626130383064363932363135
          6166396634326366323666613339646262323935346534330a626137666161376538373661393733
          62646535663135303733626336393837363233396162383864333263393939393039663465386232
          3334653638656537330a386365653664336333633134303235393134313161373936333161366636
```

После чего вывод команды также можно использовать в плейбуке:
```yaml
---
- hosts: localhost
  connection: local
  gather_facts: False
  vars:
    data: !vault |
      $ANSIBLE_VAULT;1.1;AES256
      61616430323334633331373333363932376264626439346465623139336363616161656132346164
      3330316666323662303665343261323535303738303861660a623538313730653137333364353836
      66353866646534623536326435373362386134653862383430313930343933663961666239343064
      3635633834383536320a323866396462653830623338656334353430303639353836376232643563
      6266
  tasks:
  - name: test
    debug:
      msg: "{{ data }}"
```
```console
$ ansible-playbook playbook.yaml --ask-vault-pass
Vault password:

PLAY [localhost] *****************************************************************

TASK [test] **********************************************************************
ok: [localhost] => {
    "msg": "secretData\n"
}

PLAY RECAP ***********************************************************************
localhost                  : ok=1    changed=0    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
```

## Age
Также познакомимся с утилитой [age][], которая не привязана ни к какому стеку и
позволяющая очень просто использовать ключи шифрования помимо парольной фразы
для шифрования локальных файлов.

В качестве файла возьмем `secret.yaml` из предыдущего, не забыв расшифровать его,
если он был зашифрован утилитой `ansible-vault`:
```console
$ ansible-vault decrypt secret.yaml
Vault password:
Decryption successful
$ cat secret.yaml
item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value1
  key2: secret_value2
```

### Keygen
Для генерации ключа воспользуемся командой `age-keygen` и сохраним его в файл
`key.txt`:
```console
$ age-keygen -o key.txt
Public key: age1m7dhpf204lsx6d5h56s0e96nudvjqaa2g929ec0y83tncu6umpxq7egxmp
$ age-keygen -y key.txt
age1m7dhpf204lsx6d5h56s0e96nudvjqaa2g929ec0y83tncu6umpxq7egxmp
```
Как видно при генерации ключа на экран выводится его публичная часть, которую
также можно посмотреть с помощью опции `-y` у уже созданного ключа.

### Encrypt
Для шифрования файла как раз необходима публичная часть ключа, которая указывается
в опции `-r` команды `age`:
```console
$ # сохранить зашифрованный файл в бинарном формате
$ age -r age1m7dhpf204lsx6d5h56s0e96nudvjqaa2g929ec0y83tncu6umpxq7egxmp -o secret.yaml.age secret.yaml
$ # сохранить зашифрованный файл в PEM формате
$ age -r age1m7dhpf204lsx6d5h56s0e96nudvjqaa2g929ec0y83tncu6umpxq7egxmp -a secret.yaml > secret.yaml.age
$ cat secret.yaml.age
-----BEGIN AGE ENCRYPTED FILE-----
YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBrVEVpc2dmYUgvYU1FaERO
UEtzcCtQbDlyVW9CZkVFMC9SVXlqbk5iWEhrCjVPbUdYQ2pMTFhYaWhtWkZsVmh6
Y3ZUdE0rV21yMkpIS3k0bUJSdGMzdjQKLS0tIFZrR3lwZVRud2F0ZXNuVHFHaHpz
bTE2QStkS2wvNzJwcWVRdHFaVjF5WEEKZEyOM+f8EcO8ZD3j5YLJeHffJiRW+ZCD
FBy1PpyTkZ8DyZJvY1oDeIUJRD759eOrrR933xIj6ykwUpD1UiO5cUF8KRqHRdgr
aSjIKrixWs3Y84saca+p+z67nLMSs3vpnB8TdrvKS65GWsHGQ3gvEzDnDqhMV4DC
pIRRr5xsEmtjhezQoC0h/Q2apcLKSwt5ILoZ14RMNQ==
-----END AGE ENCRYPTED FILE-----
```

### Decrypt
Для расшифровывания файла необходимо использовать ключ `-d` и `-i` с указанием
файла ключа:
```console
$ age -d -i key.txt secret.yaml.age
item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value1
  key2: secret_value2
```

### SSH Key
Для шифрования также есть возможность использовать ssh ключи, которые обычно есть
у любого разработчика:
```console
$ ssh-keygen -P '' -f $HOME/.ssh/id_rsa
Generating public/private rsa key pair.
Your identification has been saved in /home/vagrant/.ssh/id_rsa
Your public key has been saved in /home/vagrant/.ssh/id_rsa.pub
The key fingerprint is:
SHA256:CWCVKKgOUcVJ1WUp/ct8we8pVDsLQdI3dL8M8upx72k vagrant@node
The key's randomart image is:
+---[RSA 3072]----+
| o.+==oo oo.. ...|
|o ..+.. o.o. o oo|
|.. .  .  . ooo. o|
|o      . .  +.=..|
|o       S  o oo=.|
| .          =o.o.|
|           o.o..+|
|          . o..Eo|
|           .  ++ |
+----[SHA256]-----+
$ age -R $HOME/.ssh/id_rsa.pub -a secret.yaml > secret.yaml.age
$ age -d -i $HOME/.ssh/id_rsa secret.yaml.age
item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value1
  key2: secret_value2
```

### Multiple keys
А также для шифрования можно одновременно использовать несколько ключей,
комбинируя как age ключи так и ssh:
```console
$ age -r age1m7dhpf204lsx6d5h56s0e96nudvjqaa2g929ec0y83tncu6umpxq7egxmp \
  -R $HOME/.ssh/id_rsa.pub -a secret.yaml > secret.yaml.age
$ age -d -i $HOME/.ssh/id_rsa secret.yaml.age
item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value1
  key2: secret_value2
$ age -d -i key.txt secret.yaml.age
item: secret_data
list:
- secret_item1
- secret_item2
- secret_item3
dict:
  key1: secret_value1
  key2: secret_value2
```

Таким образом можно хранить файл с открытыми ключами нескольких пользователей,
например, там же где и файл с секретами. При этом каждый пользователь сможет
читать, изменять и перешифровывать такой файл без знания общей парольной фразы.

## Sops
Еще один инструмент для управления секретами - [SOPS: Secrets OPerationS][sops].
Данный инструмент использует внешнего поставщика ключей для шифрования - это
могут быть [age][] и [gpg][] как локальные утилиты, так и облачные KMS сервисы.

### GPG
Рассмотрим работу [sops][] в связке с [pgp][] ключами. Для этого с помощью
утилиты [gpg][] создадим себе ключевую пару.
```console
$ gpg --quick-gen-key --pinentry-mode=loopback --passphrase='' Alex default default 0
gpg: directory '/home/vagrant/.gnupg' created
gpg: keybox '/home/vagrant/.gnupg/pubring.kbx' created
gpg: /home/vagrant/.gnupg/trustdb.gpg: trustdb created
gpg: directory '/home/vagrant/.gnupg/openpgp-revocs.d' created
gpg: revocation certificate stored as '/home/vagrant/.gnupg/openpgp-revocs.d/4D9DC3B8A580D927DEC9D4325A66620A06DF7D54.rev'
public and secret key created and signed.

pub   rsa3072 2023-11-15 [SC]
      4D9DC3B8A580D927DEC9D4325A66620A06DF7D54
uid                      Alex
sub   rsa3072 2023-11-15 [E]
```

### Encrypt
Для шифрования с помощью [pgp][] необходимо использовать отпечаток ключа. Можно
его передавать через аргументы или переменную среды, но удобнее использовать
файл конфигурации, который также можно хранить вместе с проектом, где будут
храниться секреты. Создадим файл `.sops.yaml` в который добавим отпечаток ключа:
```yaml
creation_rules:
    - pgp: >-
        4D9DC3B8A580D927DEC9D4325A66620A06DF7D54
```

[Sops][] умеет работать со структурированными форматами файлов такие как
`json`, `yaml`, `env`, таким образом позволяя частично шифровать данные.
Тип файла по-умолчанию определяется по расширению, так что, для того чтобы
зашифровать `yaml` файл полностью, необходимо указать его тип как `binary`:
```console
$ sops -e --input-type binary secret.yaml
data: ENC[AES256_GCM,data:zpeGBft25RTki9dM9wzezpWBUstltnpsG2AM89kl9ZcL3ovj+8ugiQNMc0NaLDZo6OZjbCD6WcDrUQZTMgr8l8DIVhlQE1wcachu3tc4QLo2Ha77Sf37tXry8EgBknShtK3V1kcTgJZ/R0tahL4n1zdVr165FIk=,iv:9uvmPz5fuYc9OY+gWEmAWqUJn2Rgw1Rlf+05VbTr7Oc=,tag:ueirLeHOMfHHC1hBWOZtOg==,type:str]
sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    hc_vault: []
    age: []
    lastmodified: "2023-11-15T20:20:22Z"
    mac: ENC[AES256_GCM,data:A+eNt5DtQiISq6YEYyUwn2tG4Loxsw2AWeLgfpsffMR0u+6Ob7NECn/cRjF6Ik5Drun2CYCla0/vRbpPi1aVvyfVpxTMaLWyJJdErX6nxZ00ojqlwM1ehPJaRPSUStuTuV9uiqAMCwt3DVHwiRQUNKP0S7W1N41HC1XijfI3iNw=,iv:0YEdmAiDVDYiprZV5yyddpwqRJzLUS5aDDHsQYkThKM=,tag:OLi6Y3VKScG6lC36dAY/3w==,type:str]
    pgp:
        - created_at: "2023-11-15T20:20:22Z"
          enc: |-
            -----BEGIN PGP MESSAGE-----

            hQGMA87ESxgiPDlNAQv8DxE8QWoASICvC1P08kejqe8zdBeY0HkSbhpbPbu1CnxO
            vPkp16XFa1EyR2nSR66tRubq6PYuQeZ1yJuPL55zA80Iu7WhUNDgXNc9/aBaEMTz
            /kQ/LXfA5eN6vXP1NO7WTuh7x7BiQi+N3+HOYVYWYleckzgffpHiLIGSM5LpLPBt
            W2OK3NVw/s/hQsXv04AqpKTZCm2o97LvpytceiGylhKWd8MacM2wnHBLs+Sz9Ubq
            dMawcvipCx03FL5BFSDVbaZ4UK7BDMr/fAXF5l5EO4fIj8ghICjJN7wgWutTKUZq
            BK/dGFp/FiDZ4SLCRr+N9Eo5v45aFA3eLP71QHSfAitKKnqlQ7pER0Y9BdBVGBaj
            9v3HU3x7ebObRDaHdhuCACWOyJPiQaEHVAnH2dTrkwGEkIku89DROIFQOjEEOi96
            rGdGDRX+bY0qHUSomZ/wiQNCL2DBg/PWwxqBb5F84DJiQ9pH7McTiO0CHro9prz6
            SpCdEJ+5KPWIWTDIG55f0lwBmriqg8BjYn3O5HYHXCBgveWi9LpROcUTUJPqsjBW
            SdrWTMsVywXZUAjw9JldMLIFKyZFKTmdnV2LhxSAi8icXAWOboz+qV9RaunVPMI0
            VJ/RPTXrq/B3G+MvWQ==
            =mGt8
            -----END PGP MESSAGE-----
          fp: 4D9DC3B8A580D927DEC9D4325A66620A06DF7D54
    unencrypted_suffix: _unencrypted
    version: 3.8.1
```
Как видно зашифрованное содержимое выводится на экран по-умолчанию, для сохранения
в другой файл можно использовать опцию `--output`, либо же опцию `-i`(`--in-place`)
для изменения текущего файла. Попробуем использовать частичное шифрование
структурированного файла изменив его содержимое:
```console
$ sops -e -i secret.yaml
$ cat secret.yaml
item: ENC[AES256_GCM,data:6C16dZrteSU++Kk=,iv:SPmmwTlhDP6wkY1w4S6rosNQoE7XlG59D/WXoa0sPZg=,tag:SGwzUMqIVWOOro4oQyypyg==,type:str]
list:
    - ENC[AES256_GCM,data:PiLZhs/ED3q3dMfc,iv:XsfueJtWRfKj7Opx4pZd2ViUQtrbLIXizS8ETGtUafw=,tag:RdBMCOppAZoNLgt4daBU7w==,type:str]
    - ENC[AES256_GCM,data:D25wWO4I9drqxCTS,iv:C+mxtvYas+N7+BhCaqji4p3zodQLBYCZ617YDz4j100=,tag:+wGmRERrjnzr9hnkLTnHNw==,type:str]
    - ENC[AES256_GCM,data:SqMsPlWAtVYCsoE4,iv:Z06k7lccdL8NrdDCXjjpqfWp+/zqkzwAVfUZeFOo+E0=,tag:7e/60SQ41oYt/8xbPVE4YA==,type:str]
dict:
    key1: ENC[AES256_GCM,data:ghZWn233f1DcWTpLNw==,iv:1TnxUTA4KPVVhVmE22wqRGb2QG5h3FEwmBLZlh821z4=,tag:MkrdB1u5W5EHGPgRLjucmQ==,type:str]
    key2: ENC[AES256_GCM,data:LP7JVUQnk6NGKIeEVQ==,iv:X1YSuh/Fum2E5E7mQco9bi335WreI6DdqM17Zvksk70=,tag:PqjofaBgFZ1rEtuHAI1loA==,type:str]
sops:
...
```

### Edit
Если использовать команду `sops` передав лишь имя зашифрованного файла, то
по-умолчанию открывается текстовый редактор определенный в переменной `EDITOR`
с расшифрованным содержимым файла, который можно отредактировать и он повторно
будет зашифрован:
```console
$ EDITOR=nano sops secret.yaml
  GNU nano 7.2                /tmp/1481366927/secret.yaml *
item: secret_data
list:
    - secret_item1
    - secret_item2
    - secret_item3
dict:
    key1: secret_value3
    key2: secret_value2

                                 [ Read 8 lines ]
^G Help      ^O Write Out ^W Where Is  ^K Cut       ^T Execute   ^C Location
^X Exit      ^R Read File ^\ Replace   ^U Paste     ^J Justify   ^/ Go To Line
```

### Set/Extract
С помощью опции `--extract` вместе с опцией `-d` можно расшифровать только
нужные части файла указав путь до них:
```console
$ sops -d --extract '["item"]' secret.yaml
secret_data
$ sops -d --extract '["list"][0]' secret.yaml
secret_item1
```
А с помощью опции `--set` можно таким же образом устанавливать значения:
```console
$ sops -d --extract '["dict"]["key1"]' secret.yaml
secret_value3
$ sops --set '["dict"]["key1"] "secret_value1"' secret.yaml
$ sops -d --extract '["dict"]["key1"]' secret.yaml
secret_value1
```

Таким образом возможно удобное использование в скриптах и CI/CD пайплайнах.

### Exec File/Env
С помощью подкоманды `exec-file` можно вызвать какое-либо другое приложение,
которому передастся расшифрованный временный файл:
```console
$ sops exec-file secret.yaml 'cat {}'
item: secret_data
list:
    - secret_item1
    - secret_item2
    - secret_item3
dict:
    key1: secret_value1
    key2: secret_value2
```

Также есть возможность запустить приложение задав ему переменные среды из
зашифрованного файла:
```console
$ echo -e "VAR1=VALUE1\nVAR2=VALUE2" > .env
$ sops -e -i .env
$ sops exec-env .env 'sh -c "echo $VAR1 $VAR2"'
VALUE1 VALUE2
```

### Update keys
Для изменения ключей шифрования, например когда нужно добавить ключи новых
пользователей или удалить старых, есть подкоманда `updatekeys`. Добавим наш
age ключ в `.sops.yaml`:
```yaml
creation_rules:
- pgp: >-
    4D9DC3B8A580D927DEC9D4325A66620A06DF7D54
  age: age1m7dhpf204lsx6d5h56s0e96nudvjqaa2g929ec0y83tncu6umpxq7egxmp
```
И обновим:
```console
$ sops updatekeys secret.yaml
2023/11/15 21:32:44 Syncing keys for file /vagrant/secret.yaml
The following changes will be made to the file's groups:
Group 1
    4D9DC3B8A580D927DEC9D4325A66620A06DF7D54
+++ age1m7dhpf204lsx6d5h56s0e96nudvjqaa2g929ec0y83tncu6umpxq7egxmp
Is this okay? (y/n):y
2023/11/15 21:32:46 File /vagrant/secret.yaml synced with new keys
```

Таким образом можно управлять доступом пользователей к секретам.


[ansible]:https://docs.ansible.com/ansible/latest/index.html
[ansible-vault]:https://docs.ansible.com/ansible/latest/vault_guide/index.html
[age]:https://github.com/FiloSottile/age
[sops]:https://github.com/getsops/sops
[pgp]:https://ru.wikipedia.org/wiki/PGP
[gpg]:https://www.gnupg.org/
