# Vagrant
В данной практике рассмотрим базовое взаимодействие с [vagrant][].

## Install

### VirtualBox
Для работы с [vagrant][] нам понадобится провайдер, использующийся по-умолчанию -
[VirtualBox][].
Скачать его можно со [страницы загрузок на официальном сайте][downloads-vb] или
воспользовавшись пакетным менеджером для своей ОС:

[Windows][choco-vb-install]:
```console
$ choco install virtualbox
```

[MacOS][brew-vb-install]:
```console
$ brew install --cask virtualbox
```

[Ubuntu][apt-vb-install]:
```console
$ apt install virtualbox
```

### Vagrant
Сам [vagrant][] можно [скачать и установить с официального сайта][downloads],
где указаны инструкции для разных платформ.

[Windows][downloads]:
```console
$ choco install vagrant
```

[MacOS][downloads]:
```console
$ brew install hashicorp/tap/hashicorp-vagrant
```

[Ubuntu][downloads]:
```console
$ wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
$ echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
$ sudo apt update && sudo apt install vagrant
```

После установки можно убедиться, что [vagrant][] установлен:
```console
$ vagrant -v
Vagrant 2.3.7
```

## Init
Для инициализации проекта воспользуемся командой `vagrant init`, которая создаст
в текущей директории `Vagrantfile` с конфигурацией и комментариями к ней. Данная
команда также позволяет задать имя бокса(образа виртуальной машины):
```console
$ mkdir vagrant
$ cd vagrant
$ vagrant init ubuntu/lunar64
A `Vagrantfile` has been placed in this directory. You are now
ready to `vagrant up` your first virtual environment! Please read
the comments in the Vagrantfile as well as documentation on
`vagrantup.com` for more information on using Vagrant.
$ ls
Vagrantfile
```

После этого в директории появится `Vagrantfile`, в котором довольно подробно
описаны в виде комментариев базовые возможности конфигурации. Если же посмотреть
содержимое убрав комментарии и пустые строки, то получим следующее:
```console
$ grep -Ev '^\s*#|^$' Vagrantfile # посмотрим текущее содержимое без комментариев
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/lunar64"
end
```

## Box
К сожалению [HashiCorp][] на текущий момент ограничила доступ к своим ресурсам
из России и автоматическое скачивание бокса при запуске не будет работать.
Для установки бокса его [можно найти и вручную скачать с сайта][vagrant-search].
В данной практике будет использоваться [ubuntu/lunar64][ubuntu-box], возьмем
последнюю версию для [virtualbox][]:
```console
$ curl -LO https://app.vagrantup.com/ubuntu/boxes/lunar64/versions/20230829.0.0/providers/virtualbox.box
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   165    0   165    0     0    204      0 --:--:-- --:--:-- --:--:--   204
100   151    0   151    0     0    124      0 --:--:--  0:00:01 --:--:--     0
100  704M  100  704M    0     0  11.0M      0  0:01:03  0:01:03 --:--:-- 11.3M
$ ls
Vagrantfile  virtualbox.box
```

После скачивания добавим бокс в локальную базу [vagrant][], указав путь до файла
в команде `vagrant box add`:
```console
$ vagrant box add --name ubuntu/lunar64 virtualbox.box
==> box: Box file was not detected as metadata. Adding it directly...
==> box: Adding box 'ubuntu/lunar64' (v0) for provider:
    box: Unpacking necessary files from: file:///home/alex/infra-course/example/vagrant/virtualbox.box
==> box: Successfully added box 'ubuntu/lunar64' (v0) for 'virtualbox'!
$ rm virtualbox.box
$ vagrant box list
ubuntu/lunar64 (virtualbox, 0)
```

Другой вариант - это указать ссылку на бокс в команде `vagrant box add`, но
так как ссылка, указанная на сайте [app.vagrantup.com][vagrant-search] не прямая,
а перенаправляет на другой ресурс, то необходимо получить прямую ссылку.
Для этого можно воспользоваться командой `curl`:
```console
$ curl -ILso /dev/null https://app.vagrantup.com/ubuntu/boxes/lunar64/versions/20230829.0.0/providers/virtualbox.box -w '%{url_effective}'
https://cloud-images.ubuntu.com/lunar/current/lunar-server-cloudimg-amd64-vagrant.box
```
И использовать данную ссылку непосредственно в команде `vagrant box add`, либо сделать
подстановку команды `curl`:
```console
$ vagrant box add --name ubuntu/lunar64 $(curl -ILso /dev/null https://app.vagrantup.com/ubuntu/boxes/lunar64/versions/20230829.0.0/providers/virtualbox.box -w '%{url_effective}')
==> box: Box file was not detected as metadata. Adding it directly...
==> box: Adding box 'ubuntu/lunar64' (v0) for provider:
    box: Downloading: https://cloud-images.ubuntu.com/lunar/current/lunar-server-cloudimg-amd64-vagrant.box
==> box: Box download is resuming from prior download progress
==> box: Successfully added box 'ubuntu/lunar64' (v0) for 'virtualbox'!
$ vagrant box list
ubuntu/lunar64 (virtualbox, 0)
```

## Usage
Для управления виртуальной машиной есть ряд команд:
- `vagrant up` - создает виртуальную машину согласно описанию в `Vagrantfile`
- `vagrant halt` - останавливает виртуальную машину
- `vagrant suspend` - отправляет в сон виртуальную машину
- `vagrant resume` - пробуждает от сна виртуальную машину
- `vagrant destroy` - полностью уничтожает виртуальную машину
- `vagrant reload` - перезапускает виртуальную машину перечитывая `Vagrantfile`
- `vagrant status` - отображает статус виртуальной машины
- `vagrant global-status` - отображает глобальный статус по всем виртуальным машинам
- `vagrant port` - отображает проброшенные порты виртуальной машины
- `vagrant ssh` - подключение к терминалу виртуальной машины через ssh

Поднимем виртуальную машину из созданного `Vagrantfile`:
```console
$ grep -Ev '^\s*#|^$' Vagrantfile # посмотрим текущее содержимое без комментариев
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/lunar64"
end
$ vagrant up
Bringing machine 'default' up with 'virtualbox' provider...
==> default: Importing base box 'ubuntu/lunar64'...
==> default: Matching MAC address for NAT networking...
==> default: Setting the name of the VM: vagrant_default_1693755245554_44068
==> default: Clearing any previously set network interfaces...
==> default: Preparing network interfaces based on configuration...
    default: Adapter 1: nat
==> default: Forwarding ports...
    default: 22 (guest) => 2222 (host) (adapter 1)
==> default: Running 'pre-boot' VM customizations...
==> default: Booting VM...
==> default: Waiting for machine to boot. This may take a few minutes...
    default: SSH address: 127.0.0.1:2222
    default: SSH username: vagrant
    default: SSH auth method: private key
==> default: Machine booted and ready!
==> default: Checking for guest additions in VM...
    default: The guest additions on this VM do not match the installed version of
    default: VirtualBox! In most cases this is fine, but in rare cases it can
    default: prevent things such as shared folders from working properly. If you see
    default: shared folder errors, please make sure the guest additions within the
    default: virtual machine match the version of VirtualBox you have installed on
    default: your host and reload your VM.
    default:
    default: Guest Additions Version: 6.0.0 r127566
    default: VirtualBox Version: 7.0
==> default: Mounting shared folders...
    default: /vagrant => /home/alex/infra-course/example/vagrant
```

Попробуем подключиться:
```console
$ vagrant ssh
Welcome to Ubuntu 23.04 (GNU/Linux 6.2.0-31-generic x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

  System information as of Sun Sep  3 17:42:36 UTC 2023

  System load:  0.0               Processes:               100
  Usage of /:   4.0% of 38.70GB   Users logged in:         0
  Memory usage: 20%               IPv4 address for enp0s3: 10.0.2.15
  Swap usage:   0%


0 updates can be applied immediately.

vagrant@ubuntu-lunar:~$ uname -a
Linux ubuntu-lunar 6.2.0-31-generic #31-Ubuntu SMP PREEMPT_DYNAMIC Mon Aug 14 13:42:26 UTC 2023 x86_64 x86_64 x86_64 GNU/Linux
```

```{info}
По умолчанию vagrant синхронизирует ваш текущий каталог с каталогом `/vagrant` внутри
виртуальной машины.
```

```console
vagrant@ubuntu-lunar:~$ ls /vagrant/
Vagrantfile
vagrant@ubuntu-lunar:~$ exit
```

Проверим статус:
```console
$ vagrant status
Current machine states:

default                   running (virtualbox)

The VM is running. To stop this VM, you can run `vagrant halt` to
shut it down forcefully, or you can run `vagrant suspend` to simply
suspend the virtual machine. In either case, to restart it again,
simply run `vagrant up`.

$ vagrant global-status
id       name    provider   state   directory
----------------------------------------------------------------------------
8a72feb  default virtualbox running /home/alex/infra-course/example/vagrant

The above shows information about all known Vagrant environments
on this machine. This data is cached and may not be completely
up-to-date (use "vagrant global-status --prune" to prune invalid
entries). To interact with any of the machines, you can go to that
directory and run Vagrant, or you can use the ID directly with
Vagrant commands from any directory. For example:
"vagrant destroy 1a2b3c4d"

$ vagrant port
The forwarded ports for the machine are listed below. Please note that
these values may differ from values configured in the Vagrantfile if the
provider supports automatic port collision detection and resolution.

    22 (guest) => 2222 (host)
```

```{info}
По-умолчанию проброшен 22 порт с виртуальной машины на 2222 порт хоста,
через который мы подключаемся по ssh.
```
Пробросим дополнительно порт 80 из виртуальной машины на 8080 порт хоста, для
этого добавим в `Vagrantfile` строку:
```ruby
config.vm.network "forwarded_port", guest: 80, host: 8080
```
Данная строка уже есть в качестве примера в комментариях, если файл создавался
через команду `vagrant init`, так что достаточно раскомментировать данную строку.

Перезапустим машину с новой конфигурацией:
```console
$ grep -Ev '^\s*#|^$' Vagrantfile # посмотрим текущее содержимое без комментариев
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/lunar64"
  config.vm.network "forwarded_port", guest: 80, host: 8080
end
$ vagrant reload
==> default: Attempting graceful shutdown of VM...
==> default: Clearing any previously set forwarded ports...
==> default: Clearing any previously set network interfaces...
==> default: Preparing network interfaces based on configuration...
    default: Adapter 1: nat
==> default: Forwarding ports...
    default: 80 (guest) => 8080 (host) (adapter 1)
    default: 22 (guest) => 2222 (host) (adapter 1)
==> default: Running 'pre-boot' VM customizations...
==> default: Booting VM...
==> default: Waiting for machine to boot. This may take a few minutes...
    default: SSH address: 127.0.0.1:2222
    default: SSH username: vagrant
    default: SSH auth method: private key
==> default: Machine booted and ready!
==> default: Checking for guest additions in VM...
    default: The guest additions on this VM do not match the installed version of
    default: VirtualBox! In most cases this is fine, but in rare cases it can
    default: prevent things such as shared folders from working properly. If you see
    default: shared folder errors, please make sure the guest additions within the
    default: virtual machine match the version of VirtualBox you have installed on
    default: your host and reload your VM.
    default:
    default: Guest Additions Version: 6.0.0 r127566
    default: VirtualBox Version: 7.0
==> default: Mounting shared folders...
    default: /vagrant => /home/alex/infra-course/example/vagrant
==> default: Machine already provisioned. Run `vagrant provision` or use the `--provision`
==> default: flag to force provisioning. Provisioners marked to run always will still run.
$ vagrant port
The forwarded ports for the machine are listed below. Please note that
these values may differ from values configured in the Vagrantfile if the
provider supports automatic port collision detection and resolution.

    22 (guest) => 2222 (host)
    80 (guest) => 8080 (host)
```
Как видно после команды `vagrant reload` подхватилась новая конфигурация из `Vagrantfile`.

После работы с виртуальной машиной, чтобы она не потребляла ресурсы, нужно не забыть
выключить ее командой `vagrant halt`:
```console
$ vagrant halt
==> default: Attempting graceful shutdown of VM...
```

Если виртуальная машина и ее данные более не требуется, то можно уничтожить ее командой
`vagrant destroy`:
```console
$ vagrant destroy
    default: Are you sure you want to destroy the 'default' VM? [y/N] y
==> default: Destroying VM and associated drives...
```

## Provision
Через конфигурацию в `Vagrantfile` есть возможность подготовить виртуальную машину после
запуска. Подготовим нашу машину для запуска нашего приложения. Создадим простое
приложение, которое будет отвечать на HTTP запросы. Можете использовать свой любимый язык.

Пример на `golang`:
```golang
package main

import (
        "net/http"
)

func main() {
        http.ListenAndServe("0.0.0.0:80", http.HandlerFunc(
                func(w http.ResponseWriter, r *http.Request) {
                        w.Write([]byte("Hello!\n"))
                }),
        )
}
```

Сохраним его рядом с `Vagrantfile` в `main.go`:
```console
cat<<EOF>main.go
package main

import (
        "net/http"
)

func main() {
        http.ListenAndServe("0.0.0.0:80", http.HandlerFunc(
                func(w http.ResponseWriter, r *http.Request) {
                        w.Write([]byte("Hello!\n"))
                }),
        )
}
EOF
```

Теперь после `vagrant up` произойдет синхронизация файлов нашей директории с
директорией `/vagrant` на виртуальной машине.

Добавим `shell` provisioner в `Vagrantfile`, который установит пакет `golang`, а также
запустит наше приложение в фоновом режиме:
```ruby
config.vm.provision "shell", inline: <<-SHELL
  apt-get update
  apt-get install -y golang-go
  go run /vagrant/main.go &
SHELL
```

Запустим виртуальную машину с новой конфигурацией:
```console
$ grep -Ev '^\s*#|^$' Vagrantfile # посмотрим текущее содержимое без комментариев
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/lunar64"
  config.vm.network "forwarded_port", guest: 80, host: 8080
  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y golang-go
    go run /vagrant/main.go &
  SHELL
end

$ vagrant up
Bringing machine 'default' up with 'virtualbox' provider...
==> default: Importing base box 'ubuntu/lunar64'...
==> default: Matching MAC address for NAT networking...
==> default: Setting the name of the VM: vagrant_default_1693769872988_45712
==> default: Clearing any previously set network interfaces...
==> default: Preparing network interfaces based on configuration...
==> default: Forwarding ports...
    default: 80 (guest) => 8080 (host) (adapter 1)
    default: 22 (guest) => 2222 (host) (adapter 1)
==> default: Running 'pre-boot' VM customizations...
==> default: Booting VM...
==> default: Waiting for machine to boot. This may take a few minutes...
==> default: Machine booted and ready!
==> default: Checking for guest additions in VM...
==> default: Mounting shared folders...
==> default: Running provisioner: shell...
    default: Running: inline script
...
    default: 0 upgraded, 49 newly installed, 0 to remove and 0 not upgraded.
    default: Need to get 109 MB of archives.
    default: After this operation, 454 MB of additional disk space will be used.
    default:
    default: Running kernel seems to be up-to-date.
    default:
    default: No services need to be restarted.
    default:
    default: No containers need to be restarted.
    default:
    default: No user sessions are running outdated binaries.
```

Как видно из вывода после запуска виртуальной машины запустился скрипт, который
был задан в `config.vm.provision`. Так как в нашей конфигурации задан проброс порта
80 на порт 8080 хоста, то мы можем проверить работу приложения не заходя в виртуальную
машину:
```console
$ curl localhost:8080
Hello!
```

После того, как надобность в виртуальной машине отпала можно ее уничтожить:
```console
$ vagrant destroy
    default: Are you sure you want to destroy the 'default' VM? [y/N] y
==> default: Destroying VM and associated drives...
```

[vagrant]:https://www.vagrantup.com/
[virtualbox]:https://www.virtualbox.org/
[downloads]:https://developer.hashicorp.com/vagrant/downloads
[downloads-vb]:https://www.virtualbox.org/wiki/Downloads
[brew-vb-install]:https://formulae.brew.sh/cask/virtualbox
[choco-vb-install]:https://community.chocolatey.org/packages/virtualbox
[apt-vb-install]:https://help.ubuntu.ru/wiki/virtualbox
[hashicorp]:https://www.hashicorp.com/
[vagrant-search]:https://app.vagrantup.com/boxes/search
[ubuntu-box]:https://app.vagrantup.com/ubuntu/boxes/lunar64
