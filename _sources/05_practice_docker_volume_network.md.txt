# Docker Volume/Network
В данном практическом занятии рассматривается работа с [томами(volume)][volumes],
а также различные конфигурации [сети(network)][network].

## Volume
### Local
Для работы с локальными томами используем следующий `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "storage" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "storage"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      mkdir /data;chmod 777 /data
      echo '/data *(rw)' > /etc/exports
      export DEBIAN_FRONTEND=noninteractive
      apt-get update -q
      apt-get install -yq libnss-mdns nfs-server docker.io
      usermod -a -G docker vagrant
    SHELL
  end
end
```

Для простого монтирования локальной директории внутрь контейнера можно указать путь в
опции `-v` команды `docker run`:
```console
vagrant@storage:~$ docker run -d -v /data:/usr/share/nginx/html -p 8888:80 --name nginx nginx
Unable to find image 'nginx:latest' locally
latest: Pulling from library/nginx
a803e7c4b030: Pull complete
8b625c47d697: Pull complete
4d3239651a63: Pull complete
0f816efa513d: Pull complete
01d159b8db2f: Pull complete
5fb9a81470f3: Pull complete
9b1e1e7164db: Pull complete
Digest: sha256:32da30332506740a2f7c34d5dc70467b7f14ec67d912703568daff790ab3f755
Status: Downloaded newer image for nginx:latest
95dcf19937c35d8525c5ccf1ca527cb89903df1e4c5f80c175243da00dcb5d8b
vagrant@storage:~$ echo data > /data/index.html
vagrant@storage:~$ curl localhost:8888
data
vagrant@storage:~$ docker rm -f nginx
nginx
```

Чтобы использовать том вместо монтирования существующей директории можно создать его
явно командой [`docker volume create`][vol-create]:
```console
vagrant@storage:~$ docker volume create empty
empty
```
При создании можно указать набор меток в метаданных, которые можно потом использовать
для фильтрации:
```console
vagrant@storage:~$ docker volume create new --label test=true
new
vagrant@storage:~$ docker volume create new1 --label test=true
new1
vagrant@storage:~$ docker volume ls
DRIVER    VOLUME NAME
local     empty
local     new
local     new1
vagrant@storage:~$ docker volume ls -f label=test=true
DRIVER    VOLUME NAME
local     new
local     new1
vagrant@storage:~$ docker volume prune --filter label=test=true -af
Deleted Volumes:
new
new1

Total reclaimed space: 0B
```

Также можно указать имя несуществующего тома в опции `-v` команды `docker run`:
```console
vagrant@storage:~$ docker run -d -v html:/usr/share/nginx/html -p 8888:80 --name nginx nginx
1c9c70a664504e1e05973e11933ffadda0e7d0ac34cbc712fe2c3db5790b5abc
vagrant@storage:~$ docker volume ls
DRIVER    VOLUME NAME
local     empty
local     html
vagrant@storage:~$ curl localhost:8888
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
```

```{note}
Если мы используем пустой том и монтируем его по пути в контейнере, где уже имеются
какие-либо файлы, то данные файлы будут скопированы в этот том.
```

Для получения информации о томе можно воспользоваться командой `docker volume inspect`:
```console
vagrant@storage:~$ docker volume inspect html
[
    {
        "CreatedAt": "2023-10-04T12:40:45Z",
        "Driver": "local",
        "Labels": null,
        "Mountpoint": "/var/lib/docker/volumes/html/_data",
        "Name": "html",
        "Options": null,
        "Scope": "local"
    }
]
vagrant@storage:~$ sudo ls /var/lib/docker/volumes/html/_data
50x.html  index.html
```

С содержимым в томе можно взаимодействовать также по указанному пути:
```console
vagrant@storage:~$ sudo sh -c 'echo html > /var/lib/docker/volumes/html/_data/index.html'
vagrant@storage:~$ curl localhost:8888
html
```

Можно комбинировать монтирование томов и директорий внутрь контейнера, также можно
использовать вложенность:
```console
vagrant@storage:~$ docker run -d -v html:/usr/share/nginx/html -v /data:/usr/share/nginx/html/data -p 8888:80 --name nginx nginx
1afbbeee97533280a87480f9f051029c310548c1a33a4599449d6824227f78c2
vagrant@storage:~$ curl localhost:8888
html
vagrant@storage:~$ curl localhost:8888/data/
data
```

Содержимое томов не будет стираться после удаления контейнера, а также может
использоваться одновременно несколькими контейнерами:
```console
vagrant@storage:~$ docker rm -f nginx
nginx
vagrant@storage:~$ docker run -d -v html:/usr/share/nginx/html -p 8888:80 --name nginx1 nginx
a603c5e903d8dd4ed894d7dd5cced89e0a8018efbbaef29b19f2183ce7f0933b
vagrant@storage:~$ docker run -d -v html:/usr/share/nginx/html -p 8889:80 --name nginx2 nginx
e5c41a798a90ee298204ff38431972cf6be3ada9fc140331458bd0580ff7f5d9
vagrant@storage:~$ curl localhost:8888
html
vagrant@storage:~$ curl localhost:8889
html
vagrant@storage:~$ docker rm -f nginx1 nginx2
nginx1
nginx2
```

### Remote
Хоть по-умолчанию в качестве драйвера для создания тома можно использовать только `local`,
в `linux` использование этого драйвера подразумевает возможность использования любой
файловой системы известной ядру и монтируемой командой `mount`. [Таким образом можно
использовать сетевые файловые системы, например `NFS`.][vol-shared]
Дополним наш `Vagrantfile` двумя машинами для демонстрации и запустим их:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "storage" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "storage"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      mkdir /data;chmod 777 /data
      echo '/data *(rw)' > /etc/exports
      export DEBIAN_FRONTEND=noninteractive
      apt-get update -q
      apt-get install -yq libnss-mdns nfs-server docker.io
      usermod -a -G docker vagrant
    SHELL
  end
  config.vm.define "docker1" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "docker1"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns nfs-common docker.io
      usermod -a -G docker vagrant
    SHELL
  end
  config.vm.define "docker2" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "docker2"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns nfs-common docker.io
      usermod -a -G docker vagrant
    SHELL
  end
end
```

Создадим на машинах `docker1` и `docker2` тома, которые будут использовать сетевую
файловую систему с машины `storage` и запустим с ними контейнер `nginx` на каждой машине:
```console
$ vagrant ssh docker1
vagrant@docker1:~$ docker volume create -o type=nfs -o device=:/data -o o=addr=storage.local storage
storage
vagrant@docker1:~$ docker run -d -v storage:/usr/share/nginx/html -p 8888:80 --name nginx nginx
7d9068e47ad0e6ab70f84ec528a756f2aab14db6b55c62356923d5a0697a26c7
$ vagrant ssh docker2
vagrant@docker2:~$ docker volume create -o type=nfs -o device=:/data -o o=addr=storage.local storage
storage
vagrant@docker2:~$ docker run -d -v storage:/usr/share/nginx/html -p 8888:80 --name nginx nginx
348b13f53d1f18498265c1dcb57329ba404e43ee0ccc91b4d020638699d8434a
```

Зайдем на машину `storage` и убедимся, что контейнеры на машинах `docker1` и `docker2`
используют сетевую файловую систему:
```console
$ vagrant ssh storage
vagrant@storage:~$ curl docker1.local:8888
data
vagrant@storage:~$ curl docker2.local:8888
data
vagrant@storage:~$ echo nfs-data > /data/index.html
vagrant@storage:~$ curl docker1.local:8888
nfs-data
vagrant@storage:~$ curl docker2.local:8888
nfs-data
```

## Network
Для [управления сетевой конфигурацией][net-tutor] существует команда `docker network`,
чтобы отобразить список доступный сетей можно выполнить команду `docker network ls`:
```console
vagrant@storage:~$ docker network ls
NETWORK ID     NAME      DRIVER    SCOPE
ebe779724911   bridge    bridge    local
100b3a82985d   host      host      local
9676646e1660   none      null      local
```

### Bridge
По-умолчанию для запущенных контейнеров [используется сеть `bridge`][net-bridge],
для создание новой сети с собственной конфигурацией можно воспользоваться командой
`docker network create`. Если при создании не указать параметр `--driver`, то новая
сеть также будет типа `bridge`. Создадим новую сеть и посмотрим ее конфигурацию
командой `docker network inspect`:

```console
vagrant@storage:~$ docker network create br --subnet 10.0.0.0/24
feb2d1e6a7ef18fef4b3edf81aaac94b70410c4f7b6c1a7a60fcda8a3fd48b67
vagrant@storage:~$ docker network inspect br
[
    {
        "Name": "br",
        "Id": "feb2d1e6a7ef18fef4b3edf81aaac94b70410c4f7b6c1a7a60fcda8a3fd48b67",
        "Created": "2023-10-04T20:05:54.017090666Z",
        "Scope": "local",
        "Driver": "bridge",
        "EnableIPv6": false,
        "IPAM": {
            "Driver": "default",
            "Options": {},
            "Config": [
                {
                    "Subnet": "10.0.0.0/24"
                }
            ]
        },
        "Internal": false,
        "Attachable": false,
        "Ingress": false,
        "ConfigFrom": {
            "Network": ""
        },
        "ConfigOnly": false,
        "Containers": {},
        "Options": {},
        "Labels": {}
    }
]
```

При использовании пользовательской сети типа `bridge`, в отличии от сети используемой
по-умолчанию, также добавляется функционал разрешения имен контейнеров внутри этой сети.
Для указания сети при запуске контейнера необходимо указать опцию `--network` в команде
`docker run`, сравним запуск контейнеров в сети по-умолчанию и в нашей созданной:
```console
vagrant@storage:~$ docker run -d --name first alpine sleep inf
495e509f4f9c62a68b3501d7eabd0bd1989a8feb0736cf0cb9c415bb6c83bdd6
vagrant@storage:~$ docker run -d --name second alpine sleep inf
d7960ea8f4ed5f4006a333ca24c830fdd062beefa10d56b9e13dbc8c39fe6f3c
vagrant@storage:~$ docker exec -it first sh
# ip addr show dev eth0
56: eth0@if57: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:ac:11:00:03 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.3/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
# ping -c1 first
ping: bad address 'first'
#
vagrant@storage:~$ docker exec -it second sh
# ip addr show dev eth0
58: eth0@if59: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:ac:11:00:04 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.4/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
# ping -c1 second
ping: bad address 'second'
#
vagrant@storage:~$ docker rm -f first second
first
second

vagrant@storage:~$ docker run -d --network br --name first alpine sleep inf
2479516cf26e2af9df4e158df691695ca20d8ca92c02c35ba438ef89eb18d976
vagrant@storage:~$ docker run -d --network br --name second alpine sleep inf
a6eeefe1997e114e563a92e5930858815f147ce64f2c08b3f7dd1fabd33532d5
vagrant@storage:~$ docker exec -it first sh
# ip addr show dev eth0
60: eth0@if61: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:0a:00:00:02 brd ff:ff:ff:ff:ff:ff
    inet 10.0.0.2/24 brd 10.0.0.255 scope global eth0
       valid_lft forever preferred_lft forever
# ping -c1 first
PING first (10.0.0.2): 56 data bytes
64 bytes from 10.0.0.2: seq=0 ttl=64 time=0.029 ms

--- first ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max = 0.029/0.029/0.029 ms
# ping -c1 second
PING second (10.0.0.3): 56 data bytes
64 bytes from 10.0.0.3: seq=0 ttl=64 time=0.091 ms

--- second ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max = 0.091/0.091/0.091 ms
#
vagrant@storage:~$ docker rm -f first second
first
second
```
Как видно из вывода - при использовании созданной нами сети используется заданная нами
адресация и работает разрешение имен контейнеров.

Контейнер может быть подключен одновременно к нескольким сетям, для этого после запуска
можно воспользоваться командой `docker network connect`, а для отключения от сети
командой `docker network disconnect`:
```console
vagrant@storage:~$ docker run -d --name first alpine sleep inf
4a9da4cf9b0b959298b29da8b59a920531a9fa53769525a99d22bac58eda12bb
vagrant@storage:~$ docker exec -it first ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
68: eth0@if69: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
vagrant@storage:~$ docker network connect br first
vagrant@storage:~$ docker exec -it first ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
68: eth0@if69: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
70: eth1@if71: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:0a:00:00:02 brd ff:ff:ff:ff:ff:ff
    inet 10.0.0.2/24 brd 10.0.0.255 scope global eth1
       valid_lft forever preferred_lft forever
vagrant@storage:~$ docker network disconnect bridge first
vagrant@storage:~$ docker exec -it first ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
70: eth1@if71: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:0a:00:00:02 brd ff:ff:ff:ff:ff:ff
    inet 10.0.0.2/24 brd 10.0.0.255 scope global eth1
       valid_lft forever preferred_lft forever
```

### Host
Контейнер также можно запустить в [сети хоста][net-host], так что процессы внутри
контейнера будут прослушивать порты непосредственно на адресе хоста:
```console
vagrant@storage:~$ docker run -d -v /data:/usr/share/nginx/html --network host --name nginx nginx
88dec80591b12144824b739177b1133af526ccd4074b3d752044d1a7fe9e347b
vagrant@storage:~$ curl localhost
nfs-data
vagrant@storage:~$ docker rm -f nginx
nginx
```

### IPVLAN
Попробуем объединить сети докера из нескольких виртуальных машин воспользовавшись
[драйвером `ipvlan`][net-ipvlan]. Данный драйвер позволяет разделить один сетевой
интерфейс хоста между несколькими контейнерами, так что при создании сети нам нужно
указать `parent` интерфейс из внутренней сети, которую создает `vagrant`, для того
чтобы контейнеры могли общаться в ней между собой. При использовании по-умолчанию
virtualbox provider в vagrant для внутренней сети используется сеть 192.168.56.0/24,
имя сетевого интерфейса можно получить командой:
```console
vagrant@storage:~$ ip -br a | awk '/192.168.56/{print $1}'
enp0s8
```

Создадим на каждой машине сеть типа `ipvlan` указав для каждой машины свой `ip-range`:
```console
$ vagrant ssh storage
vagrant@storage:~$ docker network create -d ipvlan --subnet=10.1.1.0/24 --ip-range=10.1.1.0/28 -o parent=enp0s8 internal
91a0d8b2235c7851fe1f90eb7cd1924d5be3b9f631f5dc6b95407523e6d8397c
$ vagrant ssh docker1
vagrant@docker1:~$ docker network create -d ipvlan --subnet=10.1.1.0/24 --ip-range=10.1.1.16/28 -o parent=enp0s8 internal
c618b6d42b63ce1f0af54de147e2ed4c28e62f1ed9bed0e9c7101aae160e1e04
$ vagrant ssh docker2
vagrant@docker2:~$ docker network create -d ipvlan --subnet=10.1.1.0/24 --ip-range=10.1.1.32/28 -o parent=enp0s8 internal
28e646c7404ff3fc6341c19fd6c9a57512d07f8c9ca5050ce370180f8954274a
```

Запустим на машинах `docker1` и `docker2` контейнеры с `nginx` и убедимся, что
полученные контейнерами адреса находятся в заданных сетях:
```console
$ vagrant ssh docker1
vagrant@docker1:~$ docker run -d --restart=always -v storage:/usr/share/nginx/html --network internal --name nginx nginx
455186759d7c6ea8374034b7aad68067c570807a61d315c31e897b3783803f18
vagrant@docker1:~$ docker inspect nginx -f '{{json .NetworkSettings.Networks.internal.IPAddress}}'
"10.1.1.17"

$ vagrant ssh docker2
vagrant@docker2:~$ docker run -d --restart=always -v storage:/usr/share/nginx/html --network internal --name nginx nginx
1460ce30d5a8ed1267e261c29097c8970e9327088a3ba831066ffe38bffa57c8
vagrant@docker2:~$ docker inspect nginx -f '{{json .NetworkSettings.Networks.internal.IPAddress}}'
"10.1.1.33"
```

Также запустим `nginx` на машине `storage` создав конфигурацию, которая будет проксировать
запросы балансируя между контейнерами на машинах `docker1` и `docker2`:
```console
vagrant@storage:~$ cat <<EOF>>default.conf
  server {
    listen       80;
    server_name  localhost;
    location / {
        proxy_pass http://docker;
    }
}

upstream docker {
    server 10.1.1.17 fail_timeout=1s;
    server 10.1.1.33 fail_timeout=1s;
}
EOF
```

Запустим контейнер с данной конфигурацией и подключим к нашей сети:
```console
vagrant@storage:~$ docker run -d -v ./default.conf:/etc/nginx/conf.d/default.conf -p 8888:80 --name nginx nginx
vagrant@storage:~$ docker network connect internal nginx
```

Теперь мы можем проверить связность из данного контейнера до контейнеров на машинах
`docker1` и `docker2`:
```console
vagrant@storage:~$ curl localhost:8888
nfs-data
```

Даже если мы остановим машину `docker1`, то `nginx` на машине `storage` будет
направлять запросы на оставшуюся машину `docker2`:
```console
$ vagrant halt docker1
==> docker1: Attempting graceful shutdown of VM...
$ vagrant ssh storage
vagrant@storage:~$ curl localhost:8888
nfs-data
$ vagrant halt docker2
==> docker2: Attempting graceful shutdown of VM...
$ vagrant ssh storage
vagrant@storage:~$ curl localhost:8888
<html>
<head><title>502 Bad Gateway</title></head>
<body>
<center><h1>502 Bad Gateway</h1></center>
<hr><center>nginx/1.25.2</center>
</body>
</html>
$ vagrant up docker1
Bringing machine 'docker1' up with 'virtualbox' provider...
==> docker1: Machine booted and ready!
$ vagrant ssh storage
vagrant@storage:~$ curl localhost:8888
nfs-data
```

[volumes]:https://docs.docker.com/storage/volumes/
[network]:https://docs.docker.com/network/
[vol-create]:https://docs.docker.com/storage/volumes/#create-and-manage-volumes
[vol-shared]:https://docs.docker.com/storage/volumes/#share-data-between-machines
[net-tutor]:https://docs.docker.com/network/network-tutorial-standalone/
[net-bridge]:https://docs.docker.com/network/drivers/bridge/
[net-host]:https://docs.docker.com/network/drivers/host/
[net-ipvlan]:https://docs.docker.com/network/drivers/ipvlan/
