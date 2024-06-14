# Docker
В данном практическом занятии вспомним основные возможности docker клиента.

## Vagrant
Для работы с докером в независимости от платформы можно воспользоваться
следующим `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/lunar64"
  config.vm.provision "docker"
end
```

## Container lifecycle
### Run
Для простого запуска команды в контейнере достаточно запустить команду `docker run`
указав имя образа и команду. В данном примере используется образ `python:3.11.5` и
команда для просмотра версии. В некоторых образах уже имеется команда для запуска
по-умолчанию, так что нет необходимости ее указывать.
```console
$ docker run python:3.11.5-alpine python --version
Unable to find image 'python:3.11.5-alpine' locally
3.11.5-alpine: Pulling from library/python
7264a8db6415: Pull complete
66e1d5e70e42: Pull complete
0448660c92fc: Pull complete
3ce23f846e31: Pull complete
efebc2e683d2: Pull complete
Digest: sha256:5d769f990397afbb2aca24b0655e404c0f2806d268f454b052e81e39d87abf42
Status: Downloaded newer image for python:3.11.5-alpine
Python 3.11.5
```
Как видно при отсутствии образа docker скачает его.

Список запущенных контейнеров можно увидеть командой `docker ps`:
```console
$ docker ps
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
```

Так как наш контейнер после вывода строки с информацией о версии завершил
свою работу, то мы не увидим запущенных контейнеров. Чтобы увидеть
список остановленных контейнеров нужно добавить опцию `-a`:
```console
$ docker ps -a
CONTAINER ID   IMAGE                  COMMAND              CREATED          STATUS                      PORTS     NAMES
6df80e8b97b3   python:3.11.5-alpine   "python --version"   12 seconds ago   Exited (0) 12 seconds ago             modest_spence
```

Чтобы очистить завершенные контейнеры можно выполнить команду `docker container prune`:
```console
$ docker container prune
WARNING! This will remove all stopped containers.
Are you sure you want to continue? [y/N] y
Deleted Containers:
6df80e8b97b3bd430f588337dd06cddd2550b82519c8f5c67ecd47fe85913134

Total reclaimed space: 0B
```

Для того, чтобы высвобождать ресурсы контейнера сразу после его завершения можно добавить
опцию `--rm` в команде `docker run`:
```console
$ docker run --rm python:3.11.5-alpine python --version
Python 3.11.5
$ docker ps -a
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
```

Для работы в интерактивном режиме с командой, запускаемой в контейнере, необходимо
использовать две опции - `-i(interactive)` и `-t(tty)`:
```console
$ docker run --rm -it python:3.11.5-alpine python
Python 3.11.5 (main, Aug 26 2023, 00:26:34) [GCC 12.2.1 20220924] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> import os
>>> os.name
'posix'
>>> exit()
```

С помощью опции `-d` можно запустить контейнер в фоновом режиме, а с помощью
опции `--name` задать имя с которым в дальнейшем удобно будет работать:
```console
$ docker run --name test -d python:3.11.5-alpine python -m http.server 8888
1a3e73023693d79e0ea13a54d3825887e52dcafdf09a7e91a336e590ccce5462
$ docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED          STATUS          PORTS     NAMES
1a3e73023693   python:3.11.5-alpine   "python -m http.serv…"   26 seconds ago   Up 26 seconds             test
```
В ответ мы получим `id` контейнера, а сам контейнер продолжит работу в фоновом режиме.
Данная команда в контейнере запустит http сервер, который работает на порту `8888`.

### Exec
С помощью команды `docker exec` мы можем выполнить команду внутри контейнера, запустим
утилиту `wget`, которая выполнит http запрос внутри контейнера:
```console
$ docker exec test wget -qO- localhost:8888
<!DOCTYPE HTML>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Directory listing for /</title>
</head>
<body>
...
```

### Stop
Остановим контейнер командой `docker stop`:
```console
$ docker stop test
test
$ docker ps -a
CONTAINER ID   IMAGE                  COMMAND                  CREATED          STATUS                       PORTS     NAMES
4504eaafb180   python:3.11.5-alpine   "python -m http.serv…"   46 seconds ago   Exited (137) 3 seconds ago             test
```

Но после остановки данные контейнера все еще остаются в системе, для их удаления можно
воспользоваться командой `docker rm`:
```console
$ docker rm test
test
$ docker ps -a
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
```

### Run Options
Опишем простой http сервер на python, расположим его в директории с `Vagrantfile`, так
что внутри виртуальной машины он будет располагаться по пути `/vagrant/test.py`:
```python
from http.server import HTTPServer, BaseHTTPRequestHandler
from os import getenv

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        file = getenv("FILE", "none")
        self.wfile.write(f"file: {file}\n".encode())
        if file != "none":
            self.wfile.write(open(file, "rb").read())

HTTPServer(('', 8888), Handler).serve_forever()
```
Данный http сервер будет слушать порт `8888` и при наличии переменной среды `FILE` выводить
содержимое файла в ответе на `GET` запрос.

С помощью опции `-v` команды `docker run` можно смонтировать директорию внутрь контейнера:
```console
$ docker run -v /vagrant:/vagrant -d python:3.11.5-alpine python /vagrant/test.py
b1b891ee04ca0556acbbd7fd6a2669176f4d5920fbb3633f6bbde7eda2322258
```

Запустить команду внутри работающего контейнера также можно в интерактивном режиме
с опциями `-it` команды `docker exec`, например можно запустить командную оболочку:
```console
$ docker exec -it test /bin/sh
/ # wget -qO- localhost:8888
file: none
/ # ls /vagrant/
Vagrantfile  test.py
/ # env
HOSTNAME=8dcddea20192
PYTHON_PIP_VERSION=23.2.1
SHLVL=1
HOME=/root
GPG_KEY=A035C8C19219BA821ECEA86B64E628F8D684696D
PYTHON_GET_PIP_URL=https://github.com/pypa/get-pip/raw/9af82b715db434abb94a0a6f3569f43e72157346/public/get-pip.py
TERM=xterm
PATH=/usr/local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
LANG=C.UTF-8
PYTHON_VERSION=3.11.5
PYTHON_SETUPTOOLS_VERSION=65.5.1
PWD=/
PYTHON_GET_PIP_SHA256=45a2bb8bf2bb5eff16fdd00faef6f29731831c7c59bd9fc2bf1f3bed511ff1fe
/ # exit
```

Работающий контейнер также можно принудительно завершить(используя сигнал `SIGKILL`) и
высвободить ресурсы командой `docker rm -f`:
```console
$ docker rm -f test
test
```

Указать переменные среды для контейнера можно с помощью опции `-e` команды `docker run`,
а также есть возможность проброса портов опцией `-p`:
```console
$ docker run -p 8888:8888 -v /vagrant:/vagrant -e FILE=/vagrant/Vagrantfile -d --name test python:3.11.5-alpine python /vagrant/test.py
3868ed2ca1d4dde97cdac888cf595bf76293f5e81693192db8a6eb4ffda9228f
```

Убедимся, что все настройки сработали сделав запрос к приложению в контейнере с хоста:
```console
$ curl localhost:8888
file: /vagrant/Vagrantfile
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/lunar64"
  config.vm.provision "docker"
end
```

### Logs
Логи самого приложения можно посмотреть с помощью команды `docker logs`:
```console
$ docker logs test
172.17.0.1 - - [19/Sep/2023 21:59:56] "GET / HTTP/1.1" 200 -
```

### Ports
А конфигурацию портов с помощью `docker port`:
```console
$ docker port test
8888/tcp -> 0.0.0.0:8888
8888/tcp -> [::]:8888
```

### Stats
Посмотреть запущенные процессы в контейнере позволяет команда `docker top`:
```console
$ docker top test
UID   PID    PPID   C  STIME  TTY  TIME      CMD
root  11609  11587  0  21:59  ?    00:00:00  python  /vagrant/test.py
```

А статистику потребления всех контейнеров можно наблюдать командой `docker stats`:
```console
CONTAINER ID NAME CPU % MEM USAGE / LIMIT   MEM % NET I/O       BLOCK I/O   PIDS
3868ed2ca1d4 test 0.01% 11.07MiB / 952.6MiB 1.16% 1.79kB / 778B 0B / 1.98MB 1
```

## Images
При указании образов контейнеров в командах `docker` используются локально загруженные,
а при их отсутствии загружаются из общедоступного реджестри [hub.docker.com][docker-hub].
Список загруженных образов можно увидеть командой `docker images`:
```console
$ docker images
REPOSITORY   TAG             IMAGE ID       CREATED       SIZE
python       3.11.5-alpine   b9b301ab01c2   3 weeks ago   52.1MB
```

### Pull
Скачивать образа в локальное хранилище можно командой `docker pull`, скачаем образ
`registry` с помощью которого можно развернуть свой локальный реджестри в докере:
```console
$ docker pull registry:2
2: Pulling from library/registry
7264a8db6415: Already exists
c4d48a809fc2: Pull complete
88b450dec42e: Pull complete
121f958bea53: Pull complete
7417fa3c6d92: Pull complete
Digest: sha256:d5f2fb0940fe9371b6b026b9b66ad08d8ab7b0d56b6ee8d5c71cb9b45a374307
Status: Downloaded newer image for registry:2
docker.io/library/registry:2
$ docker images
REPOSITORY   TAG             IMAGE ID       CREATED       SIZE
python       3.11.5-alpine   b9b301ab01c2   3 weeks ago   52.1MB
registry     2               0030ba3d620c   6 weeks ago   24.1MB
```

Запустим собственный реджестри:
```console
$ docker run -d -p 5000:5000 --name registry registry:2
83abcb32fb48828890d5f89b6bd8eb18bdcca95928e3cebf83310a182ca83d2f
```

### Push
Для того чтобы отправить образ в наш реджестри необходимо, чтобы в имени образа было
указание на конкретный реджестри. Задать имя и тег можно командой `docker tag`,
после чего отправить командой `docker push`. Так как наш реджестри запущен локально на
порту `5000`, то в качестве имени может использоваться `localhost:5000`:
```console
$ docker tag python:3.11.5-alpine localhost:5000/python:3.11.5-alpine
$ docker push localhost:5000/python:3.11.5-alpine
The push refers to repository [localhost:5000/python]
08928985481f: Pushed
7acf52b2a13c: Pushed
ce0f4c80e9b7: Pushed
9ad60c84bfbe: Pushed
4693057ce236: Pushed
3.11.5-alpine: digest: sha256:e5d592c422d6e527cb946ae6abb1886c511a5e163d3543865f5a5b9b61c01584 size: 1368
```

### Save/Load
Образ можно выгрузить из локального хранилища в файл командой `docker save`:
```console
$ docker save python:3.11.5-alpine -o python.tar
$ ls
python.tar
```
И если потребуется, то можно перенести на другую машину и загрузить в локальное хранилище
командой `docker load`:
```console
$ docker load -i python.tar
Loaded image: python:3.11.5-alpine
```

### Build
Сборка же самих образов производится командой `docker build`. Для сборки образа
опишем простой `Dockerfile`, который добавит файл `test.py` и укажет команду запуска:
```dockerfile
FROM python:3.11.5-alpine

ADD test.py /test.py

CMD ["python", "test.py"]
```
Расположим его рядом с `Vagrantfile`, чтобы внутри виртуальной машины он находился по
пути `/vagrant/Dockerfile`. В команде `docker build` передадим опцию `-t` для указания
имени образа(если не указать тег будет использоваться latest), а также путь до
директории с контекстом(там где будет производиться сборка):
```console
$ docker build -t test /vagrant/
[+] Building 0.1s (7/7) FINISHED                                                docker:default
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 112B                                                      0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   0.0s
 => [internal] load build context                                                         0.0s
 => => transferring context: 459B                                                         0.0s
 => [1/2] FROM docker.io/library/python:3.11.5-alpine                                     0.0s
 => [2/2] ADD test.py /test.py                                                            0.0s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:41a4beea6cab781be7403620f5edd2f35f242c471dc22f4a2d7820f5235b  0.0s
 => => naming to docker.io/library/test                                                   0.0s
$ docker images
REPOSITORY              TAG             IMAGE ID       CREATED          SIZE
test                    latest          41a4beea6cab   15 seconds ago   52.1MB
localhost:5000/python   3.11.5-alpine   b9b301ab01c2   3 weeks ago      52.1MB
python                  3.11.5-alpine   b9b301ab01c2   3 weeks ago      52.1MB
registry                2               0030ba3d620c   6 weeks ago      24.1MB
```

Запустим контейнер из нашего образа, предварительно завершив контейнер `test`, если он
запущен, и заодно проверим его работу:
```console
$ docker rm -f test
test
$ docker run -p 8888:8888 -d --name test test
90f3da847ba319893fa5906b9ca7ab3988dfe04824da029f1c935bdef095aa18
$ curl localhost:8888
file: none
```

## Remote
Docker имеет клиент-серверную архитектуру, утилита `docker` является клиентом, который
по умолчанию общается с docker демоном используя unix socket. Взаимодействие может
осуществляться и по другим протоколам, самый простой - это tcp. С помощью него вы можете
взаимодействовать с удаленным docker демоном точно также как и локально утилитой `docker`.
Подготовим новую виртуальную машину в vagrant изменив конфигурацию docker демона для
взаимодействия по сети через tcp, для этого можно воспользоваться `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/lunar64"
  config.vm.network "forwarded_port", guest: 2375, host: 2375
  config.vm.provision "docker" do |d|
    d.post_install_provision "shell", inline: <<-SHELL
      systemctl cat docker.service > /etc/systemd/system/docker.service
      sed -i '/ExecStart/s#$# -H tcp://0.0.0.0:2375#' /etc/systemd/system/docker.service
      systemctl daemon-reload
      systemctl restart docker.service
    SHELL
  end
end
```

После запуска вм с данной конфигурацией на хост будет проброшен tcp порт 2375, через
который можно подключиться не заходя в виртуальную машину с помощью `docker` клиента.
Для этого достаточно установить переменную среды `DOCKER_HOST`:
```console
$ export DOCKER_HOST=localhost:2375
$ docker run -d registry:2
Unable to find image 'registry:2' locally
2: Pulling from library/registry
7264a8db6415: Pull complete
c4d48a809fc2: Pull complete
88b450dec42e: Pull complete
121f958bea53: Pull complete
7417fa3c6d92: Pull complete
Digest: sha256:d5f2fb0940fe9371b6b026b9b66ad08d8ab7b0d56b6ee8d5c71cb9b45a374307
Status: Downloaded newer image for registry:2
12e58496092d71819605aef4fb8d0ebf9cf251e5ff78bb6fd43fcb18cf77e967
$ docker ps
CONTAINER ID   IMAGE        COMMAND                  CREATED         STATUS         PORTS      NAMES
12e58496092d   registry:2   "/entrypoint.sh /etc…"   6 seconds ago   Up 5 seconds   5000/tcp   fervent_varahamihira
```

[docker-hub]:https://hub.docker.com/
