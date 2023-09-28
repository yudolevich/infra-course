## Docker

```{image} ../img/docker.svg
:width: 200px
```

### Контейнерная виртуализация
```{revealjs-fragments}
* Namespaces
* Control groups
* OverlayFS
```

### Namespaces
```{revealjs-fragments}
* Mount
* Network
* PID
* User
* IPC
* UTS
* Cgroup
```

### Namespaces
```{revealjs-code-block} console
$ ls -l /proc/$$/ns
lrwxrwxrwx 1 root root 0 Sep 28 09:50 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 root root 0 Sep 28 09:50 ipc -> 'ipc:[4026531839]'
lrwxrwxrwx 1 root root 0 Sep 28 09:50 mnt -> 'mnt:[4026531841]'
lrwxrwxrwx 1 root root 0 Sep 28 09:50 net -> 'net:[4026531840]'
lrwxrwxrwx 1 root root 0 Sep 28 09:50 pid -> 'pid:[4026531836]'
lrwxrwxrwx 1 root root 0 Sep 28 09:50 user -> 'user:[4026531837]'
lrwxrwxrwx 1 root root 0 Sep 28 09:50 uts -> 'uts:[4026531838]'
```

### Основные концепции

```{revealjs-fragments}
* Image
* Container
* Docker Daemon
* Docker Client
* Registry
```

### Архитектура

```{image} ../img/docker-arch.png
```

### Images

```{image} ../img/docker-image.png
```

### Image manifest
```{revealjs-code-block} json
---
data-line-numbers: 2|3|4-8|9-15
---
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
}
```

### Dockerfile
```{revealjs-code-block} dockerfile
FROM ubuntu:22.04
COPY . /app
RUN make /app
CMD python /app/app.py
```

### Dockerfile multi-line
```{revealjs-code-block} dockerfile
RUN apt-get update && apt-get install -y \
  bzr \
  cvs \
  git \
  mercurial \
  subversion \
  && rm -rf /var/lib/apt/lists/*
```

### Layers cache

```{image} ../img/docker-cache1.png
```

### Layers cache

```{image} ../img/docker-cache2.png
```

### Dockerfile multi-stage
```{revealjs-code-block} dockerfile
---
data-line-numbers: 1-4|6-8
---
FROM golang:1.21 as build
WORKDIR /src
COPY main.go /src/main.go
RUN go build -o /bin/hello ./main.go

FROM scratch
COPY --from=build /bin/hello /bin/hello
CMD ["/bin/hello"]
```

### Container
```{revealjs-code-block} dockerfile
$ docker run <image>
$ docker ps | head -1
CONTAINER ID  IMAGE  COMMAND  CREATED  STATUS  PORTS  NAMES
$ docker inspec <container-id>
```

### Container overlay
GraphDriver
```{revealjs-code-block} dockerfile
---
data-line-numbers: 8|3|5|4
---
{
  "Data": {
    "LowerDir": "/var/lib/docker/overlay2/475958ea3aea4a3c7c95ef8744e440f483a94ec2abde19b9a06addda6a940327-init/diff:/var/lib/docker/overlay2/39l57zdl339geagnufkc0erq3/diff",
    "MergedDir": "/var/lib/docker/overlay2/475958ea3aea4a3c7c95ef8744e440f483a94ec2abde19b9a06addda6a940327/merged",
    "UpperDir": "/var/lib/docker/overlay2/475958ea3aea4a3c7c95ef8744e440f483a94ec2abde19b9a06addda6a940327/diff",
    "WorkDir": "/var/lib/docker/overlay2/475958ea3aea4a3c7c95ef8744e440f483a94ec2abde19b9a06addda6a940327/work"
  },
  "Name": "overlay2"
}
```

### Container config
```{revealjs-code-block} dockerfile
---
data-line-numbers: 2|11-16
---
{
  "Hostname": "0adf0bbbdf2b",
  "Domainname": "",
  "User": "",
  "AttachStdin": false,
  "AttachStdout": true,
  "AttachStderr": true,
  "Tty": false,
  "OpenStdin": false,
  "StdinOnce": false,
  "Env": [
    "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
  ],
  "Cmd": [
    "/bin/hello"
  ],
  "Image": "hello:latest",
  "Volumes": null,
  "WorkingDir": "/",
  "Entrypoint": null,
  "OnBuild": null,
  "Labels": {}
}
```

### Client/Server
```{revealjs-fragments}
* Unix socket
* TCP
* SSH
* TLS(HTTPS)
```

### Client context
```{revealjs-code-block} console
---
data-line-numbers: 1-3|5-8
---
$ docker context ls
NAME        DESCRIPTION           DOCKER ENDPOINT               ERROR
default *   Current DOCKER_HOST   unix:///var/run/docker.sock

$ export DOCKER_HOST=tcp://localhost:2375
$ docker context ls
NAME        DESCRIPTION           DOCKER ENDPOINT        ERROR
default *   Current DOCKER_HOST   tcp://localhost:2375
```

### Registry
```{revealjs-fragments}
* Repository
* Tag
* Manifest
* Blob
```

### Repository
```bash
/v2/_catalog
```
```{revealjs-code-block} json
{"repositories":["main","test/hello"]}
```

### Tag
```bash
/v2/test/hello/tags/list
```
```{revealjs-code-block} json
{"name":"test/hello","tags":["0.1"]}
```

### Manifest
```bash
/v2/test/hello/manifests/0.1
```
```{revealjs-code-block} json
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
}
```

### Blob
```bash
/v2/test/hello/blobs/sha256:...
```
