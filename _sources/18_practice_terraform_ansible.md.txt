# Terraform + Ansible
В данном практическом занятии рассмотрим работу связку инструментов:
[terraform][] - для развертывания инфраструктурных ресурсов и
[ansible][] - для управления конфигурациями приложений.

## Vagrant
Для работы будем использовать следующий `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "provider1" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "provider1"
    c.vm.network "private_network", type: "dhcp"
    c.vm.network "forwarded_port", guest: 8888, host: 8888
    c.vm.network "forwarded_port", guest: 8889, host: 8889
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq docker.io libnss-mdns ansible python3-psycopg2
      usermod -a -G docker vagrant
      systemctl cat docker.service > /etc/systemd/system/docker.service
      sed -i '/ExecStart/s#$# -H tcp://0.0.0.0:2375#' /etc/systemd/system/docker.service
      systemctl daemon-reload
      systemctl restart docker.service
      curl -L https://hashicorp-releases.yandexcloud.net/terraform/1.7.3/terraform_1.7.3_linux_amd64.zip \
        | zcat > /usr/local/bin/terraform
      chmod +x /usr/local/bin/terraform
      cat > /home/vagrant/.terraformrc <<EOF
provider_installation {
    network_mirror {
        url = "https://terraform-mirror.yandexcloud.net/"
        include = ["registry.terraform.io/*/*"]
}
    direct {
        exclude = ["registry.terraform.io/*/*"]
    }
}
EOF
    SHELL
  end

  config.vm.define "provider2" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "provider2"
    c.vm.network "private_network", type: "dhcp"
    c.vm.network "forwarded_port", guest: 5432, host: 5432
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq docker.io libnss-mdns 
      usermod -a -G docker vagrant
      systemctl cat docker.service > /etc/systemd/system/docker.service
      sed -i '/ExecStart/s#$# -H tcp://0.0.0.0:2375#' /etc/systemd/system/docker.service
      systemctl daemon-reload
      systemctl restart docker.service
    SHELL
  end
end
```

Данная конфигурация развернет две виртуальные машины, которые будут
использоваться как разные [провайдеры в terraform][providers]. Все операции
будем производить на машине `provider1` в директории `/home/vagrant`, если
не указано явно другое.

## Providers
Создадим файл `main.tf`, в котором опишем конфигурацию наших провайдеров.
Так как оба провайдера будут использовать одинаковый плагин, то используем
aliases для разделения их конфигурации. Утилита `terraform` использует
встроенный механизм разрешения имен и не сможет определить адрес провайдера
на другом узле, так что воспользуемся еще одним провайдером - [external][],
который позволяет использовать [data sources][data] из вывода сторонних
утилит. Таким образом получим адрес второго узла используя системный резолвер.
```tf
terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 3.0.2"
    }
  }
}

data "external" "p2_host" {
  program = [
    "/bin/sh", "-c",
    "jq -n --arg ip $(getent hosts provider2.local | cut -f1 -d' ') '{\"ip\":$ip}'"
  ]
}

provider "docker" {
  alias = "p1"
}

provider "docker" {
  host = "tcp://${data.external.p2_host.result.ip}:2375"
  alias = "p2"
}
```

```console
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Finding kreuzwerker/docker versions matching "~> 3.0.2"...
- Finding latest version of hashicorp/external...
- Installing kreuzwerker/docker v3.0.2...
- Installed kreuzwerker/docker v3.0.2 (unauthenticated)
- Installing hashicorp/external v2.3.3...
- Installed hashicorp/external v2.3.3 (unauthenticated)

Terraform has been successfully initialized!
```

## Front
Опишем в `main.tf` ресурсы для размещения frontend приложения. В качестве веб
сервера как обычно возьмем `nginx` и опишем образ с контейнером. Для размещения
же самого приложения воспользуемся механизмом [provisioners][], который
позволяет запустить внешние инструменты для конфигурирования ресурса после
его развертывания. В качестве такого инструмента будем использовать [ansible][].
Создадим файл `index.html`:
```html
<!DOCTYPE html>
<html>
<body>

<h2>Users table:</h2>

<p id="users"></p>

<script>
var req = function() {
  var http = new XMLHttpRequest();
  http.onload = function() {
    const users = JSON.parse(this.responseText);
    let text = "<table border='1'>"
    for (let x in users) {
      text += "<tr><td>" + users[x].name + "</td>";
      text += "<td>" + users[x].email + "</td></tr>";
    }
    text += "</table>"
    document.getElementById("users").innerHTML = text;
  }

  http.open("GET", "http://localhost:8889");
  http.send();
}
setInterval(req, 1000);
</script>

</body>
</html>
```
Для доставки этого файла создадим `playbook.yaml`:
```yaml
- hosts: all
  connection: local
  become: True

  tasks:
  - ansible.builtin.file:
      path: /home/vagrant/html
      owner: 101
      group: 101
    tags: html
  - ansible.builtin.copy:
      src: /home/vagrant/index.html
      dest: /home/vagrant/html/index.html
      owner: 101
      group: 101
    tags: html
```
И опишем ресурсы в `main.tf`:
```tf
terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 3.0.2"
    }
  }
}

data "external" "p2_host" {
  program = [
    "/bin/sh", "-c",
    "jq -n --arg ip $(getent hosts provider2.local | cut -f1 -d' ') '{\"ip\":$ip}'"
  ]
}

provider "docker" {
  alias = "p1"
}

provider "docker" {
  host = "tcp://${local.p2_ip}:2375"
  alias = "p2"
}

locals {
  p2_ip = data.external.p2_host.result.ip
  front_port = 8888
}

resource "docker_image" "nginx" {
  provider = docker.p1
  name         = "nginx:alpine"
}

resource "docker_container" "front" {
  provider = docker.p1
  image = docker_image.nginx.image_id
  name  = "front"
  ports {
    internal = 80
    external = local.front_port
  }
  volumes {
    container_path = "/usr/share/nginx/html/"
    host_path = "/home/vagrant/html"
  }
  provisioner "local-exec" {
    command = "ansible-playbook -i localhost, playbook.yaml -t html"
  }
}
```
Здесь мы также добавили блок `locals` с локальными переменными для использования
в ресурсах. Теперь можем запустить `terraform apply`:
```console
$ terraform apply
data.external.p2_host: Reading...
data.external.p2_host: Read complete after 0s [id=-]

Terraform used the selected providers to generate the following execution plan.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # docker_container.front will be created
  + resource "docker_container" "front" {
      + attach                                      = false
      + bridge                                      = (known after apply)
      + command                                     = (known after apply)
      + container_logs                              = (known after apply)
      + container_read_refresh_timeout_milliseconds = 15000
      + entrypoint                                  = (known after apply)
      + env                                         = (known after apply)
      + exit_code                                   = (known after apply)
      + hostname                                    = (known after apply)
      + id                                          = (known after apply)
      + image                                       = (known after apply)
      + init                                        = (known after apply)
      + ipc_mode                                    = (known after apply)
      + log_driver                                  = (known after apply)
      + logs                                        = false
      + must_run                                    = true
      + name                                        = "front"
      + network_data                                = (known after apply)
      + read_only                                   = false
      + remove_volumes                              = true
      + restart                                     = "no"
      + rm                                          = false
      + runtime                                     = (known after apply)
      + security_opts                               = (known after apply)
      + shm_size                                    = (known after apply)
      + start                                       = true
      + stdin_open                                  = false
      + stop_signal                                 = (known after apply)
      + stop_timeout                                = (known after apply)
      + tty                                         = false
      + wait                                        = false
      + wait_timeout                                = 60

      + ports {
          + external = 8888
          + internal = 80
          + ip       = "0.0.0.0"
          + protocol = "tcp"
        }

      + volumes {
          + container_path = "/usr/share/nginx/html/"
          + host_path      = "/home/vagrant/html"
        }
    }

  # docker_image.nginx will be created
  + resource "docker_image" "nginx" {
      + id          = (known after apply)
      + image_id    = (known after apply)
      + name        = "nginx:alpine"
      + repo_digest = (known after apply)
    }

Plan: 2 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

docker_image.nginx: Creating...
docker_image.nginx: Creation complete after 6s [id=sha256:6913ed9ec8d009744018c1740879327fe2e085935b2cce7a234bf05347b670d7nginx:alpine]
docker_container.front: Creating...
docker_container.front: Provisioning with 'local-exec'...
docker_container.front (local-exec): Executing: ["/bin/sh" "-c" "ansible-playbook -i localhost, playbook.yaml -t html"]

docker_container.front (local-exec): PLAY [all] *********************************************************************

docker_container.front (local-exec): TASK [Gathering Facts] *********************************************************
docker_container.front (local-exec): ok: [localhost]

docker_container.front (local-exec): TASK [ansible.builtin.file] ****************************************************
docker_container.front (local-exec): changed: [localhost]

docker_container.front (local-exec): TASK [ansible.builtin.copy] ****************************************************
docker_container.front (local-exec): changed: [localhost]

docker_container.front (local-exec): PLAY RECAP *********************************************************************
docker_container.front (local-exec): localhost                  : ok=3    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0

docker_container.front: Creation complete after 4s [id=db1eaf886bcd42eaa359a4ffb188d3dd30da0f86985902d09cdc82adf7eef5d1]
```
Как видно, создались ресурсы образа и контейнера нашего веб сервера, после чего
с помощью плейбука был размещен `html` файл на нем. Проверим созданные ресурсы:
```console
$ docker images
REPOSITORY   TAG       IMAGE ID       CREATED      SIZE
nginx        alpine    6913ed9ec8d0   6 days ago   42.6MB
$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED         STATUS         PORTS                  NAMES
db1eaf886bcd   6913ed9ec8d0   "/docker-entrypoint.…"   3 minutes ago   Up 3 minutes   0.0.0.0:8888->80/tcp   front
$ curl localhost:8888
<!DOCTYPE html>
<html>
<body>

<h2>Users table:</h2>

<p id="users"></p>

<script>
var req = function() {
  var http = new XMLHttpRequest();
  http.onload = function() {
    const users = JSON.parse(this.responseText);
    let text = "<table border='1'>"
    for (let x in users) {
      text += "<tr><td>" + users[x].name + "</td>";
      text += "<td>" + users[x].email + "</td></tr>";
    }
    text += "</table>"
    document.getElementById("users").innerHTML = text;
  }

  http.open("GET", "http://localhost:8889");
  http.send();
}
setInterval(req, 1000);
</script>

</body>
</html>
```

## DB
Развернем теперь ресурсы для нашей базы данных, для этого нам понадобятся
образ, контейнер и том для данных. Также нам понадобится создать базу и
таблицу, которые мы будем использовать в приложении, для этого также
воспользуемся [ansible][]. Допишем наш плейбук дополнительными задачами
для конфигурирования базы, также добавим возможность получения авторизационных
данных снаружи, таким образом получим следующее содержимое `playbook.yaml`:
```yaml
- hosts: all
  connection: local
  become: True
  vars:
    user: ""
    password: ""
    pg:
      login_host: provider2.local
      login_user: "{{user}}"
      login_password: "{{password}}"
      db: postgres
  module_defaults:
    community.postgresql.postgresql_db: '{{ pg }}'
    community.postgresql.postgresql_user: '{{ pg }}'
    community.postgresql.postgresql_query: '{{ pg }}'
    community.postgresql.postgresql_table: '{{ pg }}'

  tasks:
  - ansible.builtin.file:
      path: /home/vagrant/html
      owner: 101
      group: 101
    tags: html
  - ansible.builtin.copy:
      src: /home/vagrant/index.html
      dest: /home/vagrant/html/index.html
      owner: 101
      group: 101
    tags: html
  - community.postgresql.postgresql_db:
      name: app
    tags: sql
  - community.postgresql.postgresql_user:
      name: app
      db: app
      priv: ALL
    tags: sql
  - community.postgresql.postgresql_table:
      db: app
      name: users
      columns:
      - id serial primary key
      - name varchar(50)
      - email varchar(100)
    tags: sql
```
Далее опишем новые ресурсы в `main.tf`:

```tf
terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.2"
    }
  }
}

data "external" "p2_host" {
  program = [
    "/bin/sh", "-c",
    "jq -n --arg ip $(getent hosts provider2.local | cut -f1 -d' ') '{\"ip\":$ip}'"
  ]
}

provider "docker" {
  alias = "p1"
}

provider "docker" {
  host  = "tcp://${local.p2_ip}:2375"
  alias = "p2"
}

locals {
  p2_ip      = data.external.p2_host.result.ip
  front_port = 8888
  back_port  = 8889
  db_port    = 5432
}

variable "db_user" {
  type = string
}

variable "db_password" {
  type      = string
  sensitive = true
}

resource "docker_image" "nginx" {
  provider = docker.p1
  name     = "nginx:alpine"
}

resource "docker_container" "front" {
  provider = docker.p1
  image    = docker_image.nginx.image_id
  name     = "front"
  ports {
    internal = 80
    external = local.front_port
  }
  volumes {
    container_path = "/usr/share/nginx/html/"
    host_path      = "/home/vagrant/html"
  }
  provisioner "local-exec" {
    command = "ansible-playbook -i localhost, playbook.yaml -t html"
  }
}

resource "docker_image" "postgres" {
  provider = docker.p2
  name     = "postgres:alpine"
}

resource "docker_volume" "data" {
  provider = docker.p2
  name     = "data"
}

resource "docker_container" "db" {
  provider = docker.p2
  image    = docker_image.postgres.image_id
  name     = "db"
  env = [
    "POSTGRES_USER=${var.db_user}",
    "POSTGRES_PASSWORD=${var.db_password}"
  ]
  ports {
    internal = local.db_port
    external = local.db_port
  }
  volumes {
    volume_name    = resource.docker_volume.data.name
    container_path = "/var/lib/postgresql/data"
  }
  provisioner "local-exec" {
    command = "ansible-playbook -i localhost, playbook.yaml -t sql -e user=${var.db_user} -e password=${var.db_password}"
  }
}
```

Дополним локальные переменные, а также воспользуемся [input variables][variables]
для описания входных переменных для указания пользователя и пароля
для подключения к субд при запуске. После чего применим конфигурацию, указав
в качестве пользователя и пароля `postgres`:
```console
$ terraform apply --auto-approve
var.db_password
  Enter a value:

var.db_user
  Enter a value: postgres

data.external.p2_host: Reading...
docker_image.nginx: Refreshing state... [id=sha256:6913ed9ec8d009744018c1740879327fe2e085935b2cce7a234bf05347b670d7nginx:alpine]
docker_container.front: Refreshing state... [id=db1eaf886bcd42eaa359a4ffb188d3dd30da0f86985902d09cdc82adf7eef5d1]
data.external.p2_host: Read complete after 0s [id=-]

Terraform used the selected providers to generate the following execution plan.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # docker_container.db will be created
  + resource "docker_container" "db" {
      + attach                                      = false
      + bridge                                      = (known after apply)
      + command                                     = (known after apply)
      + container_logs                              = (known after apply)
      + container_read_refresh_timeout_milliseconds = 15000
      + entrypoint                                  = (known after apply)
      + env                                         = (sensitive value)
      + exit_code                                   = (known after apply)
      + hostname                                    = (known after apply)
      + id                                          = (known after apply)
      + image                                       = (known after apply)
      + init                                        = (known after apply)
      + ipc_mode                                    = (known after apply)
      + log_driver                                  = (known after apply)
      + logs                                        = false
      + must_run                                    = true
      + name                                        = "db"
      + network_data                                = (known after apply)
      + read_only                                   = false
      + remove_volumes                              = true
      + restart                                     = "no"
      + rm                                          = false
      + runtime                                     = (known after apply)
      + security_opts                               = (known after apply)
      + shm_size                                    = (known after apply)
      + start                                       = true
      + stdin_open                                  = false
      + stop_signal                                 = (known after apply)
      + stop_timeout                                = (known after apply)
      + tty                                         = false
      + wait                                        = false
      + wait_timeout                                = 60

      + ports {
          + external = 5432
          + internal = 5432
          + ip       = "0.0.0.0"
          + protocol = "tcp"
        }

      + volumes {
          + container_path = "/var/lib/postgresql/data"
          + volume_name    = "data"
        }
    }

  # docker_image.postgres will be created
  + resource "docker_image" "postgres" {
      + id          = (known after apply)
      + image_id    = (known after apply)
      + name        = "postgres:alpine"
      + repo_digest = (known after apply)
    }

  # docker_volume.data will be created
  + resource "docker_volume" "data" {
      + driver     = (known after apply)
      + id         = (known after apply)
      + mountpoint = (known after apply)
      + name       = "data"
    }

Plan: 3 to add, 0 to change, 0 to destroy.
docker_volume.data: Creating...
docker_image.postgres: Creating...
docker_volume.data: Creation complete after 0s [id=data]
docker_image.postgres: Still creating... [10s elapsed]
docker_image.postgres: Creation complete after 14s [id=sha256:09ac24c200ca00e1e699bd76aea3987400e6451b98171ca6d648f0c8f637c23epostgres:alpine]
docker_container.db: Creating...
docker_container.db: Provisioning with 'local-exec'...
docker_container.db (local-exec): (output suppressed due to sensitive value in config)
docker_container.db: Creation complete after 4s [id=4d236481a7b1e25e5e370be94ddb3495dece6ebe5409de4a4ad4214db9b7a0fa]

Apply complete! Resources: 3 added, 0 changed, 0 destroyed.
```
Зайдем на машину второго провайдера `provider2` и проверим созданные ресурсы
и конфигурацию базы:
```console
$ docker images
REPOSITORY   TAG       IMAGE ID       CREATED       SIZE
postgres     alpine    09ac24c200ca   12 days ago   243MB
$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED         STATUS         PORTS                    NAMES
4d236481a7b1   09ac24c200ca   "docker-entrypoint.s…"   3 minutes ago   Up 3 minutes   0.0.0.0:5432->5432/tcp   db
$ docker volume ls
DRIVER    VOLUME NAME
local     data
$ docker exec -it db su postgres -c 'psql app'
psql (16.2)
Type "help" for help.

app=# \d users
                                    Table "public.users"
 Column |          Type          | Collation | Nullable |              Default
--------+------------------------+-----------+----------+-----------------------------------
 id     | integer                |           | not null | nextval('users_id_seq'::regclass)
 name   | character varying(50)  |           |          |
 email  | character varying(100) |           |          |
Indexes:
    "users_pkey" PRIMARY KEY, btree (id)

app=# \du
                             List of roles
 Role name |                         Attributes
-----------+------------------------------------------------------------
 app       |
 postgres  | Superuser, Create role, Create DB, Replication, Bypass RLS
```

## Back
Осталось развернуть ресурсы backend приложения, чтобы связать frontend и базу
данных. Для этого потребуется собрать образ приложения, опишем его в `main.go`:
```golang
package main

import (
        "context"
        "encoding/json"
        "fmt"
        "net/http"
        "os"

        "github.com/jackc/pgx/v5"
)

type users struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
}

func main() {
        ctx := context.Background()
        conn, err := pgx.Connect(ctx, string(os.Getenv("CONNECTION")))
        if err != nil {
                fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
                os.Exit(1)
        }
        defer conn.Close(ctx)

        http.ListenAndServe("0.0.0.0:80", http.HandlerFunc(
                func(w http.ResponseWriter, r *http.Request) {
                        rows, err := conn.Query(ctx, "select * from users")
                        if err != nil {
                                fmt.Printf("error db query: %s", err)
                                return
                        }

                        users, err := pgx.CollectRows(rows, pgx.RowToStructByName[users])
                        if err != nil {
                                fmt.Printf("error collect rows: %s", err)
                                return
                        }

                        jsonUsers, err := json.Marshal(users)
                        if err != nil {
                                fmt.Printf("error marshal json: %s", err)
                        }

                        w.Header().Add("Access-Control-Allow-Origin", "*")
                        w.Write(jsonUsers)
                }),
        )
}
```
А также `Dockerfile`:
```dockerfile
FROM golang:1.21 as build

WORKDIR /src

COPY main.go /src/main.go
RUN go mod init example \
  && go mod tidy \
  && CGO_ENABLED=0 go build -o /bin/app ./main.go

FROM scratch
COPY --from=build /bin/app /app
CMD ["/app"]
```
Данное приложение будет принимать строку соединения из переменной среды
`CONNECTION` для подключения к субд и отдавать в `json` формате содержимое
нашей таблицы по запросу.

Теперь мы можем описать ресурсы образа со сборкой и контейнера в `main.tf`:
```tf
terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 3.0.2"
    }
  }
}

data "external" "p2_host" {
  program = [
    "/bin/sh", "-c",
    "jq -n --arg ip $(getent hosts provider2.local | cut -f1 -d' ') '{\"ip\":$ip}'"
  ]
}

provider "docker" {
  alias = "p1"
}

provider "docker" {
  host = "tcp://${local.p2_ip}:2375"
  alias = "p2"
}

locals {
  p2_ip = data.external.p2_host.result.ip
  front_port = 8888
  back_port = 8889
  db_port = 5432
}

variable "db_user" {
  type = string
}

variable "db_password" {
  type = string
  sensitive = true
}

resource "docker_image" "nginx" {
  provider = docker.p1
  name         = "nginx:alpine"
}

resource "docker_container" "front" {
  provider = docker.p1
  image = docker_image.nginx.image_id
  name  = "front"
  ports {
    internal = 80
    external = local.front_port
  }
  volumes {
    container_path = "/usr/share/nginx/html/"
    host_path = "/home/vagrant/html"
  }
  provisioner "local-exec" {
    command = "ansible-playbook -i localhost, playbook.yaml -t html"
  }
}

resource "docker_image" "back" {
  provider = docker.p1
  name = "back"
  build {
    context = "."
  }
}

resource "docker_container" "back" {
  provider = docker.p1
  image = docker_image.back.image_id
  name  = "back"
  ports {
    internal = 80
    external = local.back_port
  }
  env = [
    "CONNECTION=postgres://${var.db_user}:${var.db_password}@${local.p2_ip}:${local.db_port}/app?sslmode=disable"
  ]
  depends_on = [
    docker_container.db
  ]
}

resource "docker_image" "postgres" {
  provider = docker.p2
  name         = "postgres:alpine"
}

resource "docker_volume" "data" {
  provider = docker.p2
  name = "data"
}

resource "docker_container" "db" {
  provider = docker.p2
  image = docker_image.postgres.image_id
  name  = "db"
  env = [
    "POSTGRES_USER=${var.db_user}",
    "POSTGRES_PASSWORD=${var.db_password}"
  ]
  ports {
    internal = local.db_port
    external = local.db_port
  }
  volumes {
    volume_name = resource.docker_volume.data.name
    container_path = "/var/lib/postgresql/data"
  }
  provisioner "local-exec" {
    command = "ansible-playbook -i localhost, playbook.yaml -t sql -e user=${var.db_user} -e password=${var.db_password}"
  }
}
```

Передавать переменные в `terraform` можно через аргументы запуска, через файл
или через переменные среды, добавив префикс `TF_VAR`. Применим обновленную
конфигурацию:
```console
$ export TF_VAR_db_user=postgres TF_VAR_db_password=postgres
$ terraform apply --auto-approve
docker_image.nginx: Refreshing state... [id=sha256:6913ed9ec8d009744018c1740879327fe2e085935b2cce7a234bf05347b670d7nginx:alpine]
data.external.p2_host: Reading...
docker_container.front: Refreshing state... [id=db1eaf886bcd42eaa359a4ffb188d3dd30da0f86985902d09cdc82adf7eef5d1]
data.external.p2_host: Read complete after 0s [id=-]
docker_volume.data: Refreshing state... [id=data]
docker_image.postgres: Refreshing state... [id=sha256:09ac24c200ca00e1e699bd76aea3987400e6451b98171ca6d648f0c8f637c23epostgres:alpine]
docker_container.db: Refreshing state... [id=4d236481a7b1e25e5e370be94ddb3495dece6ebe5409de4a4ad4214db9b7a0fa]

Terraform used the selected providers to generate the following execution plan.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # docker_container.back will be created
  + resource "docker_container" "back" {
      + attach                                      = false
      + bridge                                      = (known after apply)
      + command                                     = (known after apply)
      + container_logs                              = (known after apply)
      + container_read_refresh_timeout_milliseconds = 15000
      + entrypoint                                  = (known after apply)
      + env                                         = (sensitive value)
      + exit_code                                   = (known after apply)
      + hostname                                    = (known after apply)
      + id                                          = (known after apply)
      + image                                       = (known after apply)
      + init                                        = (known after apply)
      + ipc_mode                                    = (known after apply)
      + log_driver                                  = (known after apply)
      + logs                                        = false
      + must_run                                    = true
      + name                                        = "back"
      + network_data                                = (known after apply)
      + read_only                                   = false
      + remove_volumes                              = true
      + restart                                     = "no"
      + rm                                          = false
      + runtime                                     = (known after apply)
      + security_opts                               = (known after apply)
      + shm_size                                    = (known after apply)
      + start                                       = true
      + stdin_open                                  = false
      + stop_signal                                 = (known after apply)
      + stop_timeout                                = (known after apply)
      + tty                                         = false
      + wait                                        = false
      + wait_timeout                                = 60

      + ports {
          + external = 8889
          + internal = 80
          + ip       = "0.0.0.0"
          + protocol = "tcp"
        }
    }

  # docker_image.back will be created
  + resource "docker_image" "back" {
      + id          = (known after apply)
      + image_id    = (known after apply)
      + name        = "back"
      + repo_digest = (known after apply)

      + build {
          + cache_from   = []
          + context      = "."
          + dockerfile   = "Dockerfile"
          + extra_hosts  = []
          + remove       = true
          + security_opt = []
          + tag          = []
        }
    }

Plan: 2 to add, 0 to change, 0 to destroy.
docker_image.back: Creating...
docker_image.back: Still creating... [10s elapsed]
docker_image.back: Still creating... [1m0s elapsed]
docker_image.back: Creation complete after 1m0s [id=sha256:2649c141424a9d53a8bd3a03121455597710408550d2722f7a0df0e1626fc271back]
docker_container.back: Creating...
docker_container.back: Creation complete after 0s [id=597f469d9d32783b9839e9da187c7cf6af0ecf68270f2580eee9ef38e0708adc]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
```

## Result
После развертывания всех ресурсов и их конфигурирования мы получили работающее
приложение, которое доступно по адресу [localhost:8888](http://localhost:8888/):

![](img/terraform-ansible1.png)

Добавим в базу пару строк и убедимся, что они отразились в приложении. Для этого
на втором узле `provider2` выполним:
```console
vagrant@provider2:~$ docker exec -it db su postgres -c 'psql app'
psql (16.2)
Type "help" for help.

app=# insert into users (name,email) values ('alex','alex@mail.ru');
INSERT 0 1
app=# insert into users (name,email) values ('vasya','vasya@gmail.com');
INSERT 0 1
```
После чего увидим:

![](img/terraform-ansible2.png)

После того, как необходимости в ресурсах больше нет - их следует уничтожить
командами `terraform destroy` и `vagrant destroy`.

По итогу мы получили работающее приложение. Таким образом связка инструментов
[terraform][] и [ansible][] позволяет разделить ответственность между
управлением развертыванием инфраструктурных ресурсов и управлением
конфигурациями компонентов, упрощая в целом развертывание конечных приложений.


[terraform]:https://developer.hashicorp.com/terraform
[ansible]:https://docs.ansible.com/ansible/latest/index.html
[providers]:https://developer.hashicorp.com/terraform/language/providers
[data]:https://developer.hashicorp.com/terraform/language/data-sources
[external]:https://github.com/hashicorp/terraform-provider-external
[provisioners]:https://developer.hashicorp.com/terraform/language/resources/provisioners/syntax
[variables]:https://developer.hashicorp.com/terraform/language/values/variables
