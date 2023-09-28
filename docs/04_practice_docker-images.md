# Docker Images
В данном практическом занятии рассматриваются операции с docker образами:
директивы Dockerfile, сборка, работа с registry.

## Vagrant
Для работы с докером в независимости от платформы можно воспользоваться
следующим `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/lunar64"
  config.vm.provision "docker"
end
```

## Dockerfile
[Рассмотрим различные директивы, которые можно использовать в Dockerfile][dockerfile].
Во всех примерах `Dockerfile` будет находиться в директории проекта
вместе с `Vagrantfile` и внутри виртуальной машины будет доступен по пути
`/vagrant/Dockerfile`.

### Hello
Создадим простой [Dockerfile][], который содержит две директивы -
[`FROM`][from] для указания базового образа и [`CMD`][cmd] для команды, которая запустится
при старте контейнера:
```dockerfile
FROM python:3.11.5-alpine
CMD ["python", "-c", "print('hello')"]
```
Соберем образ с именем `hello` из данного [Dockerfile][], при простом запуске увидим
как отработала команда в директиве [`CMD`][cmd]:
```console
$ docker build -t hello /vagrant/ # без указания тега образ будет иметь тег latest
[+] Building 4.5s (5/5) FINISHED                                                docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 102B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   1.9s
 => [1/1] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  2.6s
 => => resolve docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  0.0s
 => => sha256:66e1d5e70e420aa86a23bd8b4eebf2a6eb60b4aff9ee8a6ca52e27 622.31kB / 622.31kB  0.3s
 => => sha256:0448660c92fc1b6557cf48081d99028eeaba7809a1c2f23018e2331f 12.46MB / 12.46MB  1.8s
 => => sha256:5d769f990397afbb2aca24b0655e404c0f2806d268f454b052e81e39d8 1.65kB / 1.65kB  0.0s
 => => sha256:e5d592c422d6e527cb946ae6abb1886c511a5e163d3543865f5a5b9b61 1.37kB / 1.37kB  0.0s
 => => sha256:b9b301ab01c2af0b5f52069b97e1d89885433dc55b6623ad6aa4a43b2e 6.26kB / 6.26kB  0.0s
 => => sha256:7264a8db6415046d36d16ba98b79778e18accee6ffa71850405994cffa 3.40MB / 3.40MB  0.7s
 => => sha256:3ce23f846e315e35618c4604547bbe6aa2b48a0601c335c76b8fe02729bb5c 241B / 241B  0.8s
 => => extracting sha256:7264a8db6415046d36d16ba98b79778e18accee6ffa71850405994cffa9be7d  0.1s
 => => sha256:efebc2e683d297ae71acae9230d17af8b95684d9a4f9b7f601a8e1e9b1 3.11MB / 3.11MB  1.5s
 => => extracting sha256:66e1d5e70e420aa86a23bd8b4eebf2a6eb60b4aff9ee8a6ca52e27f51f57b1b  0.2s
 => => extracting sha256:0448660c92fc1b6557cf48081d99028eeaba7809a1c2f23018e2331f3cfe4b2  0.4s
 => => extracting sha256:3ce23f846e315e35618c4604547bbe6aa2b48a0601c335c76b8fe02729bb5c4  0.0s
 => => extracting sha256:efebc2e683d297ae71acae9230d17af8b95684d9a4f9b7f601a8e1e9b103bda  0.2s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:caaa3b6bf89788867da2a2d8fbd3396b3486c6bf7922b2177e6c2dabd8f6  0.0s
 => => naming to docker.io/library/hello                                                  0.0s
$ docker run --rm hello:latest
hello
```

Для получения детальной информации об образе можно воспользоваться командой
`docker inspect`:
```console
$ docker image inspect hello:latest
[
    {
        "Id": "sha256:b82039dd53d96094d9e8a5f0e5698fc26590f029b239883d931ce8a753772b71",
        "RepoTags": [
            "hello:latest"
        ],
...
```

### Copy/Add
Для копирования файлов внутрь образа из контекста сборки можно воспользоваться
инструкциями [`COPY`][copy] и [`ADD`][add]. Контекст сборки указывается в команде
`docker build` в виде пути(в нашем случае это путь `/vagrant/`.

Возьмем скрипт на python, который будет отвечать на HTTP запросы по порту 8888 и расположим
его рядом с [Dockerfile][] под именем `main.py`:
```python
#!/usr/bin/env python

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
Опишем [Dockerfile][] следующим образом:
```dockerfile
FROM python:3.11.5-alpine

COPY main.py /main.py
ADD https://example.com /example.html

CMD ["python", "main.py"]
```
В данном примере инструкция `COPY` скопирует файл из контекста сборки(`/vagrant/`) внутрь
образа, а `ADD` скачает файл пор заданному URL и сохранит внутри образа с именем
`example.html`. После чего мы задаем команду запуска интерпретатора python, которому
передаем на исполнение наш файл.
```{note}
Основное отличие `COPY` от `ADD` - это возоможность `ADD` скачивать файлы по URL, а также
автоматически разархивировать tar архивы. Но рекомендуется использовать `COPY`, так у
данной директивы более очевидное поведение, а скачивание и разархивирование производить
в директивах `RUN`.
```

Соберем образ с именем `main`:
```console
$ docker build -t main /vagrant/
[+] Building 1.6s (9/9) FINISHED                                                docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 151B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   1.0s
 => [internal] load build context                                                         0.0s
 => => transferring context: 29B                                                          0.0s
 => https://example.com                                                                   0.5s
 => [1/3] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  0.0s
 => CACHED [2/3] COPY main.py /main.py                                                    0.0s
 => CACHED [3/3] ADD https://example.com /example.html                                    0.0s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:b82039dd53d96094d9e8a5f0e5698fc26590f029b239883d931ce8a75377  0.0s
 => => naming to docker.io/library/main                                                   0.0s
```

И проверим его работу:
```console
$ docker run -d -p 8888:8888 --name main main:latest
f74c1f74e7bc09a0797a7717c144ffa64c5412f9621c66dd7d822057375ddd65
$ curl localhost:8888
file: none
```
Как видно наше приложение запустилось и обрабатывает HTTP запросы.

Также в инструкция [`COPY`][copy] и [`ADD`][add] есть возможность задать права файла
внутри образа с помощью опции `--chmod` и владельца опцией `--chown`. Дадим файлу `main.py`
права на запуск, так что можно будет в [`CMD`][cmd] указать не интерпретатор python, а наш
исполняемый файл:
```dockerfile
FROM python:3.11.5-alpine

COPY --chmod=555 main.py /main.py
ADD https://example.com /example.html

CMD ["/main.py"]
```

```console
$ docker build -t main /vagrant/
[+] Building 1.5s (9/9) FINISHED                                                docker:default
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 154B                                                      0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   1.0s
 => [internal] load build context                                                         0.0s
 => => transferring context: 29B                                                          0.0s
 => https://example.com                                                                   0.5s
 => [1/3] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  0.0s
 => CACHED [2/3] COPY --chmod=555 main.py /main.py                                        0.0s
 => CACHED [3/3] ADD https://example.com /example.html                                    0.0s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:09368b064a4bbde6e44a79f7743bbd361d1f82ddcdf8b239f504a085997a  0.0s
 => => naming to docker.io/library/main                                                   0.0s
```

### Env/Arg
Для передачи параметров во время сборки служит инструкция [`ARG`][arg], с помощью нее
можно задать переменные среды, которые будут существовать только на время сборки.
Если же необходимо задать переменные среды, которые будут существовать также при запуске
контейнера, то можно воспользоваться директивой [`ENV`][env].

В данном примере мы задаем переменную `FILE` как аргумент сборки, а также передаем во
время сборки в инструкцию [`ENV`][env], чтобы данная переменная была доступна также и после
сборки во время запуска контейнера:
```dockerfile
FROM python:3.11.5-alpine

ARG FILE
ENV FILE="${FILE}"

COPY --chmod=555 main.py /main.py
ADD https://example.com /example.html

CMD ["/main.py"]
```

Передача аргументов при запуске сборки осуществляется опцией `--build-arg`:
```console
$ docker build -t main --build-arg FILE=/example.html /vagrant/
[+] Building 0.7s (9/9) FINISHED                                                docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 185B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   0.5s
 => [internal] load build context                                                         0.0s
 => => transferring context: 29B                                                          0.0s
 => https://example.com                                                                   0.1s
 => [1/3] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  0.0s
 => CACHED [2/3] COPY --chmod=555 main.py /main.py                                        0.0s
 => CACHED [3/3] ADD https://example.com /example.html                                    0.0s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:e90a9bec6dacd27430cd93793704b5f1c7abcd668232d19cdbf94ec60a05  0.0s
 => => naming to docker.io/library/main                                                   0.0s
$ docker rm -f main
main
$ docker run -d -p 8888:8888 --name main main:latest
c5e2af35ebb1fa9c5df54a6c4e699a8e8944df3ae0b22b587d79e5a3deb78fcc
$ curl -s localhost:8888 | grep title
    <title>Example Domain</title>
```

### Run
Для запуска команд внутри образа во время сборки есть инструкция [`RUN`][run].
С помощью нее вы можете подготовить образ используя утилиты, которые находятся
в базовом образе. Перепишем файл `main.py` с использованием внешних зависимостей, которые
необходимо будет установить во время сборки:
```python
#!/usr/bin/env python
from os import getenv
from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()

@app.get("/", response_class=HTMLResponse)
def read_example():
    file = getenv("FILE", "none")
    ret = f"file: {file}\n".encode()
    if file != "none":
        ret += open(file, "rb").read()
    return ret
```

Соответственно [Dockerfile][] может выглядеть так:
```dockerfile
FROM python:3.11.5-alpine

ENV FILE="/example.html"

COPY main.py .
ADD https://example.com ${FILE}
RUN pip install fastapi "uvicorn[standard]"

CMD ["uvicorn", "main:app", "--host=0.0.0.0", "--port=8888"]
```

Соберем и проверим:
```console
$ docker build -t main /vagrant/
[+] Building 10.8s (10/10) FINISHED                                             docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 245B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   0.9s
 => [internal] load build context                                                         0.0s
 => => transferring context: 29B                                                          0.0s
 => CACHED [1/4] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca  0.0s
 => CACHED https://example.com                                                            0.5s
 => [2/4] COPY main.py .                                                                  0.0s
 => [3/4] ADD https://example.com /example.html                                           0.0s
 => [4/4] RUN pip install fastapi "uvicorn[standard]"                                     8.8s
 => exporting to image                                                                    0.6s
 => => exporting layers                                                                   0.6s
 => => writing image sha256:a872628e15170dc4f697d8c403387f4f29ae495f11365f7343c5c008b552  0.0s
 => => naming to docker.io/library/main                                                   0.0s
$ docker rm -f main
main
$ docker run -d -p 8888:8888 --name main main:latest
959fb49fe6eb777f76279ee9e779c09f3d01ff94a9a505ddc8186e559fc4ed2d
$ curl -s localhost:8888 | head -5
file: /example.html
<!doctype html>
<html>
<head>
    <title>Example Domain</title>
```

### Cache
Так как docker в процессе сборки кэширует слои, то важна последовательность сборки.
При изменении вышележащего слоя потребуется пересборка всех последующих. Если изменится
файл `main.py`, то потребуется заново скачать и установить зависимости:
```console
$ echo >> /vagrant/main.py
$ docker build -t main /vagrant/
[+] Building 11.2s (10/10) FINISHED                                             docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 245B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   0.4s
 => [internal] load build context                                                         0.0s
 => => transferring context: 381B                                                         0.0s
 => CACHED https://example.com                                                            1.4s
 => CACHED [1/4] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca  0.0s
 => [2/4] COPY main.py .                                                                  0.0s
 => [3/4] ADD https://example.com /example.html                                           0.0s
 => [4/4] RUN pip install fastapi "uvicorn[standard]"                                     8.8s
 => exporting to image                                                                    0.5s
 => => exporting layers                                                                   0.5s
 => => writing image sha256:b0d51dd802f65df529cded6d35e920b2cd58e99d9248c132443b79c6777b  0.0s
 => => naming to docker.io/library/main                                                   0.0s
```

Зададим установку зависимостей в нижележащем слое в [Dockerfile][]:
```dockerfile
FROM python:3.11.5-alpine

ENV FILE="/example.html"

RUN pip install fastapi "uvicorn[standard]"
COPY main.py .
ADD https://example.com ${FILE}

CMD ["uvicorn", "main:app", "--host=0.0.0.0", "--port=8888"]
```

Теперь последующие сборки не потребуют выполнять скачивание и установку зависимостей,
что значительно ускорит их время выполнения:
```console
$ docker build -t main /vagrant/
[+] Building 1.0s (10/10) FINISHED                                              docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 245B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   0.5s
 => [1/4] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  0.0s
 => CACHED https://example.com                                                            0.5s
 => [internal] load build context                                                         0.0s
 => => transferring context: 29B                                                          0.0s
 => CACHED [2/4] RUN pip install fastapi "uvicorn[standard]"                              0.0s
 => [3/4] COPY main.py .                                                                  0.0s
 => [4/4] ADD https://example.com /example.html                                           0.0s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:4cf84d3406df7701d0c143deb30c913e4d88f68353740e03684c3056e83d  0.0s
 => => naming to docker.io/library/main                                                   0.0s
$ echo >> /vagrant/main.py
$ docker build -t main /vagrant/
[+] Building 0.6s (10/10) FINISHED                                              docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 245B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   0.4s
 => [1/4] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  0.0s
 => [internal] load build context                                                         0.0s
 => => transferring context: 382B                                                         0.0s
 => CACHED https://example.com                                                            0.1s
 => CACHED [2/4] RUN pip install fastapi "uvicorn[standard]"                              0.0s
 => [3/4] COPY main.py .                                                                  0.0s
 => [4/4] ADD https://example.com /example.html                                           0.0s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:0657e4aef1a267c34f131d0c3e26b82823abd18db063c24b623201edd8b0  0.0s
 => => naming to docker.io/library/main                                                   0.0s
```

### Expose
Для указания портов, которые могут использоваться для подключения к приложению
в контейнере используется инструкция [`EXPOSE`][expose]:
```dockerfile
FROM python:3.11.5-alpine

ENV FILE="/example.html"

RUN pip install fastapi "uvicorn[standard]"
COPY main.py .
ADD https://example.com ${FILE}

EXPOSE 8888

CMD ["uvicorn", "main:app", "--host=0.0.0.0", "--port=8888"]
```

После сборки образа можно наблюдать выставленные порты в команде `docker image inspect`,
а также после запуска, если явно не задан проброс портов, в команде `docker ps`:
```console
$ docker build -t main /vagrant/
[+] Building 1.4s (10/10) FINISHED                                              docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 258B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   0.9s
 => [1/4] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  0.0s
 => https://example.com                                                                   0.5s
 => [internal] load build context                                                         0.0s
 => => transferring context: 29B                                                          0.0s
 => CACHED [2/4] RUN pip install fastapi "uvicorn[standard]"                              0.0s
 => CACHED [3/4] COPY main.py .                                                           0.0s
 => CACHED [4/4] ADD https://example.com /example.html                                    0.0s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:c34234caa5431df50019ca07f4072639dcfa51a737f175c2654b228706a6  0.0s
 => => naming to docker.io/library/main                                                   0.0s
$ docker image inspect main:latest --format='{{json .Config.ExposedPorts}}'
{"8888/tcp":{}}
$ docker rm -f main
main
$ docker run -d --name main main:latest
91a4683a36bc635e4ea91531f5a9861a7d95572eb30b17883123c66c9836bd5e
$ docker ps
CONTAINER ID   IMAGE         COMMAND                  CREATED         STATUS         PORTS      NAMES
91a4683a36bc   main:latest   "uvicorn main:app --…"   5 seconds ago   Up 4 seconds   8888/tcp   main
```

### Label
С помощью инструкции [`LABEL`][label] можно добавить метаданные в образ в виде
`<key>=<value>`, которые могут служить как дополнительная информация об образе или
использоваться каким-либо способом в сборочных конвейерах:
```dockerfile
FROM python:3.11.5-alpine

ENV FILE="/example.html"

RUN pip install fastapi "uvicorn[standard]"
COPY main.py .
ADD https://example.com ${FILE}

EXPOSE 8888

LABEL version="0.1"

CMD ["uvicorn", "main:app", "--host=0.0.0.0", "--port=8888"]
```

После сборки значения директив [`LABEL`][label] можно также посмотреть командой
`docker image inspect`:
```console
$ docker build -t main /vagrant/
[+] Building 1.4s (10/10) FINISHED                                              docker:default
 => [internal] load build definition from Dockerfile                                      0.0s
 => => transferring dockerfile: 279B                                                      0.0s
 => [internal] load .dockerignore                                                         0.0s
 => => transferring context: 2B                                                           0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                   0.9s
 => [1/4] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b0655  0.0s
 => https://example.com                                                                   0.5s
 => [internal] load build context                                                         0.0s
 => => transferring context: 29B                                                          0.0s
 => CACHED [2/4] RUN pip install fastapi "uvicorn[standard]"                              0.0s
 => CACHED [3/4] COPY main.py .                                                           0.0s
 => CACHED [4/4] ADD https://example.com /example.html                                    0.0s
 => exporting to image                                                                    0.0s
 => => exporting layers                                                                   0.0s
 => => writing image sha256:f08ac61ac1576278ff220148791ae1fe212c60aa2efd0cb70e725f708f2a  0.0s
 => => naming to docker.io/library/main                                                   0.0s
$ docker image inspect main:latest --format='{{json .Config.Labels}}'
{"version":"0.1"}
```

### Cmd/Entrypoint
Для указания команды, которая будет запускаться при старте контейнера есть две директивы:
[`CMD`][cmd] и [`ENTRYPOINT`][entrypoint]. Если в [Dockerfile][] указана только одна из
них, то она и будет определять команду запуска. Если же указаны обе, то команда запуска
будет строиться сначала из параметров в [`ENTRYPOINT`][entrypoint], затем из параметров
в [`CMD`][cmd]. [Подробнее о их взаимодействии можно почитать в документации.][cmd-entry]

Разделим команду запуска на две инструкции:
```dockerfile
FROM python:3.11.5-alpine

ENV FILE="/example.html"

RUN pip install fastapi "uvicorn[standard]"
COPY main.py .
ADD https://example.com ${FILE}

EXPOSE 8888

LABEL version="0.1"

ENTRYPOINT ["uvicorn", "main:app"]
CMD ["--host=0.0.0.0", "--port=8888"]
```

После сборки можно убедиться, что контейнер функционирует как обычно:
```console
$ docker build -t main /vagrant/
[+] Building 1.6s (10/10) FINISHED                                             docker:default
 => [internal] load build definition from Dockerfile                                     0.0s
 => => transferring dockerfile: 291B                                                     0.0s
 => [internal] load .dockerignore                                                        0.0s
 => => transferring context: 2B                                                          0.0s
 => [internal] load metadata for docker.io/library/python:3.11.5-alpine                  1.1s
 => [1/4] FROM docker.io/library/python:3.11.5-alpine@sha256:5d769f990397afbb2aca24b065  0.0s
 => https://example.com                                                                  0.5s
 => [internal] load build context                                                        0.0s
 => => transferring context: 29B                                                         0.0s
 => CACHED [2/4] RUN pip install fastapi "uvicorn[standard]"                             0.0s
 => CACHED [3/4] COPY main.py .                                                          0.0s
 => CACHED [4/4] ADD https://example.com /example.html                                   0.0s
 => exporting to image                                                                   0.0s
 => => exporting layers                                                                  0.0s
 => => writing image sha256:c5e6676b703a475f3a09b547771d7d8afb8df158f92bee7b2734dc717f6  0.0s
 => => naming to docker.io/library/main                                                  0.0s
$ docker rm -f main
main
$ docker run -d -p 8888:8888 --name main main:latest
9e7f7d23812ed19f077e9ac29eff54e57b4efbcfa67b8b311acced68254a2ce7
$ curl -s localhost:8888 | head -5
file: /example.html
<!doctype html>
<html>
<head>
    <title>Example Domain</title>
```

Теперь же мы можем переопределить параметры в [`CMD`][cmd], указав их при запуске
контейнера в конце команды `docker run`, а также [`ENTRYPOINT`][entrypoint] опцией
`--entrypoint`:
```console
$ docker rm -f main
main
$ docker run -d -p 8889:8889 --name main main:latest --host=0.0.0.0 --port=8889
acbc469c530e7c7fcc2e7e0032e2e84000a6dac38c4a82014002e7f5fcaf2806
$ curl -s localhost:8889 | head -5
file: /example.html
<!doctype html>
<html>
<head>
    <title>Example Domain</title>
$ docker rm -f main
main
$ docker run --rm --name main --entrypoint echo main:latest hello
hello
```

### Multi-Stage
Docker позволяет производить [multi-stage][] сборки, которые позволяют оптимизировать
результирующий образ, вложив в него только необходимое для запуска приложения, а
зависимости, необходимые для сборки приложения, иметь в отдельном стейдже только на
время сборки.

Возьмем [Dockerfile][] со следующим содержимым:
```dockerfile
FROM golang:1.21 as build
WORKDIR /src
COPY <<EOF /src/main.go
package main

import "fmt"

func main() {
  fmt.Println("hello, world")
}
EOF
RUN go build -o /bin/hello ./main.go

FROM scratch
COPY --from=build /bin/hello /bin/hello
CMD ["/bin/hello"]
```
В данном примере сборка образа происходит на основе базового образа `golang:1.21`, а
после итоговый бинарный файл копируется в пустой образ `FROM scratch`.

Соберем образ и запустим:
```console
$ docker build -t hello /vagrant/
[+] Building 38.5s (10/10) FINISHED                                            docker:default
 => [internal] load build definition from Dockerfile                                     0.0s
 => => transferring dockerfile: 290B                                                     0.0s
 => [internal] load .dockerignore                                                        0.0s
 => => transferring context: 2B                                                          0.0s
 => [internal] load metadata for docker.io/library/golang:1.21                           0.4s
 => [build 1/4] FROM docker.io/library/golang:1.21@sha256:19600fdcae402165dcdab18cb964  32.1s
 => => resolve docker.io/library/golang:1.21@sha256:19600fdcae402165dcdab18cb9649540bde  0.0s
 => => sha256:19600fdcae402165dcdab18cb9649540bde6be7274dedb5d205b2f840 2.36kB / 2.36kB  0.0s
 => => sha256:b47a222d28fa95680198398973d0a29b82a968f03e7ef361cc8ded5 24.03MB / 24.03MB  8.0s
 => => sha256:debce5f9f3a9709885f7f2ad3cf41f036a3b57b406b27ba3a88392 64.11MB / 64.11MB  13.5s
 => => sha256:b17c35044f4062d83c815434615997eed97697daae8745c6dd39dc367 1.58kB / 1.58kB  0.0s
 => => sha256:2159148dcc081245165b2aa99fc5a94ca9818bece66839d8eb11c9335 7.22kB / 7.22kB  0.0s
 => => sha256:167b8a53ca4504bc6aa3182e336fa96f4ef76875d158c1933d3e2f 49.56MB / 49.56MB  12.2s
 => => sha256:91b457aaf04f424db4f223ea7aad4b196d4a62da58d6f45938233e 92.30MB / 92.30MB  25.0s
 => => extracting sha256:167b8a53ca4504bc6aa3182e336fa96f4ef76875d158c1933d3e2fa19c57e0  2.3s
 => => sha256:b0ed6cc9b50977796e8eb9b270ad9c62922003c0090aa3e5ec26a1 66.99MB / 66.99MB  24.6s
 => => sha256:92b30a24413a45c9744b79bafa4c7717eafff586a9332210abbd384f7778 155B / 155B  13.7s
 => => extracting sha256:b47a222d28fa95680198398973d0a29b82a968f03e7ef361cc8ded562e4d84  0.7s
 => => extracting sha256:debce5f9f3a9709885f7f2ad3cf41f036a3b57b406b27ba3a8839283157870  2.9s
 => => extracting sha256:91b457aaf04f424db4f223ea7aad4b196d4a62da58d6f45938233e0f54bd16  2.5s
 => => extracting sha256:b0ed6cc9b50977796e8eb9b270ad9c62922003c0090aa3e5ec26a165cfcb9c  4.3s
 => => extracting sha256:92b30a24413a45c9744b79bafa4c7717eafff586a9332210abbd384f77785d  0.0s
 => [internal] preparing inline document                                                 0.0s
 => [build 2/4] WORKDIR /src                                                             0.1s
 => [build 3/4] COPY <<EOF /src/main.go                                                  0.0s
 => [build 4/4] RUN go build -o /bin/hello ./main.go                                     5.6s
 => [stage-1 1/1] COPY --from=build /bin/hello /bin/hello                                0.0s
 => exporting to image                                                                   0.0s
 => => exporting layers                                                                  0.0s
 => => writing image sha256:643d97e6291d22876100fff8e6edc90eccd8b045011cd8a4aa95524d8d5  0.0s
 => => naming to docker.io/library/hello                                                 0.0s
$ docker run --rm hello:latest
hello, world
```

После сборки можно увидеть, что итоговый образ состоит из одного слоя, а его размер
менее двух мегабайт:
```console
$ docker image inspect hello:latest --format='{{json .RootFS}}'
{"Type":"layers","Layers":["sha256:ce0d1c9c6c2209b264099514fad8f2d036136f31628db20370847bc3fcd68393"]}
$ docker images hello
REPOSITORY   TAG       IMAGE ID       CREATED              SIZE
hello        latest    643d97e6291d   About a minute ago   1.8MB
```

## Registry
Для хранения образов обычно используется специальный реестр(registry), с которым docker
клиент может взаимодействовать по http протоколу и который имеет свое [API][]. По-умолчанию
docker клиент использует публичный реестр [hub.docker.com][docker-hub].

### Run
Поднимем свой локальный реестр из образа `registry:2`, который по-умолчанию
слушает порт 5000:
```console
$ docker run -d -p 5000:5000 --name registry registry:2
Unable to find image 'registry:2' locally
2: Pulling from library/registry
7264a8db6415: Already exists
c4d48a809fc2: Pull complete
88b450dec42e: Pull complete
121f958bea53: Pull complete
7417fa3c6d92: Pull complete
Digest: sha256:d5f2fb0940fe9371b6b026b9b66ad08d8ab7b0d56b6ee8d5c71cb9b45a374307
Status: Downloaded newer image for registry:2
6d6ce64e7f7302e287fe4e8325f48adc4c939b6e91400318c52528ceceefad09
```
Проверить его работоспособность можно командой `curl`:
```console
$ curl localhost:5000/v2/
{}
```

### Push
Для отправки образа в registry необходимо задать ему имя, в котором будет указание на
конкретный registry. Для этого можно воспользоваться командой `docker tag`:
```console
$ docker tag hello:latest localhost:5000/hello # без указания тега будет использован latest
$ docker tag hello:latest localhost:5000/hello:0.1 # можно указать тег явно
$ docker tag main:latest localhost:5000/main
$ docker tag main:latest localhost:5000/main:python3.11.5
```

После чего можно отправить образы в локальный реестр командой `docker push`:
```console
$ docker push localhost:5000/hello
Using default tag: latest
The push refers to repository [localhost:5000/hello]
ce0d1c9c6c22: Pushed
latest: digest: sha256:3f8d6a1d510560a70b02dcc9722501c15fc4a1a78e11333913b863017dde40d7 size: 527
$ docker push localhost:5000/hello:0.1
The push refers to repository [localhost:5000/hello]
ce0d1c9c6c22: Layer already exists
0.1: digest: sha256:3f8d6a1d510560a70b02dcc9722501c15fc4a1a78e11333913b863017dde40d7 size: 527
$ docker push localhost:5000/main
Using default tag: latest
The push refers to repository [localhost:5000/main]
18e02cddb804: Pushed
3a5eb5dbf933: Pushed
c4b567ddc865: Pushed
08928985481f: Pushed
7acf52b2a13c: Pushed
ce0f4c80e9b7: Pushed
9ad60c84bfbe: Pushed
4693057ce236: Pushed
latest: digest: sha256:4957a623702236127336a3e54a5dad314c14e411fbd3096e5f669919afa90d21 size: 1994
$ docker push localhost:5000/main:
latest        python3.11.5
$ docker push localhost:5000/main:python3.11.5
The push refers to repository [localhost:5000/main]
18e02cddb804: Layer already exists
3a5eb5dbf933: Layer already exists
c4b567ddc865: Layer already exists
08928985481f: Layer already exists
7acf52b2a13c: Layer already exists
ce0f4c80e9b7: Layer already exists
9ad60c84bfbe: Layer already exists
4693057ce236: Layer already exists
python3.11.5: digest: sha256:4957a623702236127336a3e54a5dad314c14e411fbd3096e5f669919afa90d21 size: 1994
```

Как видно при отправке образа, который отличается только тегом, слои не отправляются
повторно.

### Catalog
Список репозиториев в реестре можно получить через [api][], например отправив запрос
командой `curl`:
```console
$ curl localhost:5000/v2/_catalog
{"repositories":["hello","main"]}
```

### Tags/Manifests
Список тегов в каждом репозитории можно получить следующим запросом:
```console
$ curl localhost:5000/v2/hello/tags/list
{"name":"hello","tags":["latest","0.1"]}
$ curl localhost:5000/v2/main/tags/list
{"name":"main","tags":["latest","python3.11.5"]}
```

По тегу же можно получить манифест образа - описание состава образа в `json` формате,
конкретный манифест можно получить запросом:
```console
$ curl localhost:5000/v2/hello/manifests/0.1
{
   "schemaVersion": 1,
   "name": "hello",
   "tag": "0.1",
   "architecture": "amd64",
   "fsLayers": [
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:e00c2c33f044a4c9b2ce7fd84a33c63bb82fdf88dcc17994a14ecbf0dd0be01d"
      }
   ],
   "history": [
...
```

Также с помощью заголовка `Accept` можно выбрать версию манифеста:
```console
$ curl -v localhost:5000/v2/hello/manifests/0.1 -H 'Accept: application/vnd.docker.distribution.manifest.v2+json'
*   Trying 127.0.0.1:5000...
* Connected to localhost (127.0.0.1) port 5000 (#0)
> GET /v2/hello/manifests/0.1 HTTP/1.1
> Host: localhost:5000
> User-Agent: curl/7.88.1
> Accept: application/vnd.docker.distribution.manifest.v2+json
>
< HTTP/1.1 200 OK
< Content-Length: 527
< Content-Type: application/vnd.docker.distribution.manifest.v2+json
< Docker-Content-Digest: sha256:3f8d6a1d510560a70b02dcc9722501c15fc4a1a78e11333913b863017dde40d7
< Docker-Distribution-Api-Version: registry/2.0
< Etag: "sha256:3f8d6a1d510560a70b02dcc9722501c15fc4a1a78e11333913b863017dde40d7"
< X-Content-Type-Options: nosniff
<
{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 633,
      "digest": "sha256:643d97e6291d22876100fff8e6edc90eccd8b045011cd8a4aa95524d8d5a711f"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 1113611,
         "digest": "sha256:e00c2c33f044a4c9b2ce7fd84a33c63bb82fdf88dcc17994a14ecbf0dd0be01d"
      }
   ]
```

[dockerfile]:https://docs.docker.com/engine/reference/builder/
[from]:https://docs.docker.com/engine/reference/builder/#from
[cmd]:https://docs.docker.com/engine/reference/builder/#cmd
[copy]:https://docs.docker.com/engine/reference/builder/#copy
[add]:https://docs.docker.com/engine/reference/builder/#add
[env]:https://docs.docker.com/engine/reference/builder/#env
[arg]:https://docs.docker.com/engine/reference/builder/#arg
[run]:https://docs.docker.com/engine/reference/builder/#run
[expose]:https://docs.docker.com/engine/reference/builder/#expose
[label]:https://docs.docker.com/engine/reference/builder/#label
[entrypoint]:https://docs.docker.com/engine/reference/builder/#entrypoint
[cmd-entry]:https://docs.docker.com/engine/reference/builder/#understand-how-cmd-and-entrypoint-interact
[multi-stage]:https://docs.docker.com/build/building/multi-stage/
[api]:https://docs.docker.com/registry/spec/api/
[docker-hub]:https://hub.docker.com/
