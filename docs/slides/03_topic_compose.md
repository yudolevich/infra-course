## Network/Volume + Compose

```{image} ../img/compose.png
:width: 200px
```

### Network

### Publish ports
```console
## -p, --publish ip:[hostPort]:containerPort
$ docker run -p 8080:80
$ docker run -p 192.168.1.100:8080:80
$ docker run -p 8080:80/udp
$ docker run -p 8080:80/tcp -p 8080:80/udp
```

### Network resolve
```console
$ docker run --dns
$ docker run --hostname
$ docker run --add-host
```

### Network
```{image} ../img/docker-net1.png
```

### Network inspect
```console
$ docker network inspect bridge
```
```{revealjs-code-block} json
---
data-line-numbers: 3-16|19-27|28-35
---
[
  {
   "Name": "bridge",
   "Id": "f7ab26d71dbd6f557852c7156ae0574bbf62c42f539b50c8ebde0f728a253b6f",
   "Scope": "local",
   "Driver": "bridge",
   "EnableIPv6": false,
   "IPAM": {
    "Driver": "default",
    "Options": null,
    "Config": [
     {
      "Subnet": "172.17.0.1/16",
      "Gateway": "172.17.0.1"
     }
    ]
   },
   "Internal": false,
   "Containers": {
    "3386a527aa08b37ea9232cbcace2d2458d49f44bb05a6b775fba7ddd40d8f92c": {
     "Name": "networktest",
     "EndpointID": "647c12443e91faf0fd508b6edfe59c30b642abb60dfab890b4bdccee38750bc1",
     "MacAddress": "02:42:ac:11:00:02",
     "IPv4Address": "172.17.0.2/16",
     "IPv6Address": ""
    }
   },
   "Options": {
    "com.docker.network.bridge.default_bridge": "true",
    "com.docker.network.bridge.enable_icc": "true",
    "com.docker.network.bridge.enable_ip_masquerade": "true",
    "com.docker.network.bridge.host_binding_ipv4": "0.0.0.0",
    "com.docker.network.bridge.name": "docker0",
    "com.docker.network.driver.mtu": "9001"
   },
   "Labels": {}
  }
]
```

### Manage networks
```console
$ docker network create
$ docker network rm
$ docker network ls
$ docker network prune
```

### Select network
```console
$ docker run --network
$ docker network connect
$ docker network disconnect
```

### Multiple networks
```{image} ../img/docker-net2.png
```

### Network drivers
```{revealjs-fragments}
* bridge
* host
* overlay
* ipvlan
* macvlan
* none
```

### Volume
```{image} ../img/docker-volume1.png
```

### Volume
```
-v|--volume=[HOST-DIR:]CONTAINER-DIR[:OPTIONS]

--mount type=TYPE,TYPE-SPECIFIC-OPTION[,...]

  type=bind,source=/path/on/host,destination=/path/in/container

  type=volume,source=my-volume,destination=/path/in/container

  type=tmpfs,tmpfs-size=512M,destination=/path/in/container
```

### Volume manage
```{revealjs-code-block} console
---
data-line-numbers: 1|2-3|4-14|15
---
$ docker volume create my-vol
$ docker volume ls
local               my-vol
$ docker volume inspect my-vol
[
    {
        "Driver": "local",
        "Labels": {},
        "Mountpoint": "/var/lib/docker/volumes/my-vol/_data",
        "Name": "my-vol",
        "Options": {},
        "Scope": "local"
    }
]
$ docker volume rm my-vol
```

### Shared volume
```{image} ../img/docker-volume2.svg
```

### Volume driver
```{revealjs-code-block} console
---
data-line-numbers: 1-6|8-12|13-17
---
$ docker volume create \
	--driver local \
	--opt type=cifs \
	--opt device=//uxxxxx.your-server.de/backup \
	--opt o=addr=uxxxxx.your-server.de,username=uxxxxxxx,password=*****,file_mode=0777,dir_mode=0777 \
	--name cif-volume

$ docker plugin install --grant-all-permissions vieux/sshfs
$ docker volume create --driver vieux/sshfs \
  -o sshcmd=test@node2:/home/test \
  -o password=testpassword \
  sshvolume
$ docker run -d \
  --name sshfs-container \
  --volume-driver vieux/sshfs \
  --mount src=sshvolume,target=/app,volume-opt=sshcmd=test@node2:/home/test,volume-opt=password=testpassword \
  nginx:latest
```

### Compose
```{revealjs-fragments}
* docker-compose standalone
* docker compose plugin
```

### Compose
```{image} ../img/compose1.png
```

### Compose file
```{revealjs-code-block} yaml
---
data-line-numbers: 1|2|14|21|27|31|35|2-12|14-19
---
services:
  frontend:
    image: example/webapp
    ports:
      - "443:8043"
    networks:
      - front-tier
      - back-tier
    configs:
      - httpd-config
    secrets:
      - server-certificate

  backend:
    image: example/database
    volumes:
      - db-data:/etc/data
    networks:
      - back-tier

volumes:
  db-data:
    driver: flocker
    driver_opts:
      size: "10GiB"

configs:
  httpd-config:
    external: true

secrets:
  server-certificate:
    external: true

networks:
  # The presence of these objects is sufficient to define them
  front-tier: {}
  back-tier: {}
```
