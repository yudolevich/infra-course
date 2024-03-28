# Terraform
Данное практическое занятие посвящено основам управления инфраструктурой
с использованием [terraform][]. В качестве инфраструктурного [провайдера
будет использоваться docker][docker-provider].

## Vagrant
Для работы с [terraform][] воспользуемся следующим `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "node" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.provision "shell", inline: <<-SHELL
apt-get update -q
apt-get install -yq docker.io
usermod -a -G docker vagrant
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
end
```

## Init

Состояние инфраструктуры, управляемое с помощью [terraform][], описывается
с помощью [языка конфигурации terraform][language].
Опишем состояние ресурсов, которые мы хотим получить в файле `main.tf`:
```tf
terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 3.0.2"
    }
  }
}

provider "docker" {}

resource "docker_image" "nginx" {
  name         = "nginx:latest"
}

resource "docker_container" "nginx" {
  image = docker_image.nginx.image_id
  name  = "tutorial"
  ports {
    internal = 80
    external = 8000
  }
}
```

В данной конфигурации мы указываем провайдера ресурсов -
[docker][docker-provider], образ и описание контейнера, которые мы хотим
получить. Подробности о конфигурации ресурсов можно получить
[в документации провайдера][docker-provider]. С помощью команды
`terraform validate` можно проверить описанную конфигурацию:
```console
$ terraform validate
Success! The configuration is valid.
```

После описания конфигурации запустим команду `terraform init`
для инициализации рабочей директории:

```console
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Finding kreuzwerker/docker versions matching "~> 3.0.2"...
- Installing kreuzwerker/docker v3.0.2...
- Installed kreuzwerker/docker v3.0.2 (unauthenticated)

Terraform has created a lock file .terraform.lock.hcl to record the provider
selections it made above. Include this file in your version control repository
so that Terraform can guarantee to make the same selections by default when
you run "terraform init" in the future.

Terraform has been successfully initialized!
```
Как видно, данная команда также установила [необходимый плагин провайдера
][docker-provider] и внесла информацию о нем в файл
`.terraform.lock.hcl`.


## Plan
Для просмотра вносимых изменений в инфраструктуру, описанных в
конфигурации, до непосредственного изменения существует команда
`terraform plan`:
```console
$ terraform plan

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # docker_container.nginx will be created
  + resource "docker_container" "nginx" {
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
      + name                                        = "tutorial"
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
          + external = 8000
          + internal = 80
          + ip       = "0.0.0.0"
          + protocol = "tcp"
        }
    }

  # docker_image.nginx will be created
  + resource "docker_image" "nginx" {
      + id          = (known after apply)
      + image_id    = (known after apply)
      + name        = "nginx:latest"
      + repo_digest = (known after apply)
    }

Plan: 2 to add, 0 to change, 0 to destroy.
```
Данная команда выводит все изменения, которые произойдут с ресурсами при
применении данной конфигурации.

## Apply
Применить текущую конфигурацию можно командой `terraform apply`:
```console
$ terraform apply --auto-approve

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # docker_container.nginx will be created
  + resource "docker_container" "nginx" {
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
      + name                                        = "tutorial"
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
          + external = 8000
          + internal = 80
          + ip       = "0.0.0.0"
          + protocol = "tcp"
        }
    }

  # docker_image.nginx will be created
  + resource "docker_image" "nginx" {
      + id          = (known after apply)
      + image_id    = (known after apply)
      + name        = "nginx:latest"
      + repo_digest = (known after apply)
    }

Plan: 2 to add, 0 to change, 0 to destroy.
docker_image.nginx: Creating...
docker_image.nginx: Still creating... [10s elapsed]
docker_image.nginx: Creation complete after 10s [id=sha256:247f7abff9f7097bbdab57df76fedd124d1e24a6ec4944fb5ef0ad128997ce05nginx:latest]
docker_container.nginx: Creating...
docker_container.nginx: Creation complete after 0s [id=6c830b003efd1939725e78490013225267a625a429780d1c1b918e6ac3cf7650]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
```
После выполнения данной команды мы получим инфраструктурные ресурсы в
том состоянии, в котором они описаны в файле `main.tf`, а также файл
состояния `terraform.tfstate`, в котором [terraform][] будет хранить
текущее состояние инфраструктуры:
```console
$ ls
main.tf  terraform.tfstate
$ docker images
REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        latest    247f7abff9f7   3 months ago   187MB
$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED         STATUS         PORTS                  NAMES
6c830b003efd   247f7abff9f7   "/docker-entrypoint.…"   5 minutes ago   Up 5 minutes   0.0.0.0:8000->80/tcp   tutorial
$ curl -I localhost:8000
HTTP/1.1 200 OK
```

## Change
Внесем изменения в конфигурационный файл `main.tf`, например, изменив
тег образа и порт контейнера:
```tf
terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 3.0.2"
    }
  }
}

provider "docker" {}

resource "docker_image" "nginx" {
  name         = "nginx:alpine"
}

resource "docker_container" "nginx" {
  image = docker_image.nginx.image_id
  name  = "tutorial"
  ports {
    internal = 80
    external = 8001
  }
}
```
И попробуем применить данную конфигурацию. Команда `terraform apply` без
аргументов покажет вносимые изменения и запросит подтверждение перед
применением:
```console
$ terraform apply
docker_image.nginx: Refreshing state... [id=sha256:247f7abff9f7097bbdab57df76fedd124d1e24a6ec4944fb5ef0ad128997ce05nginx:latest]
docker_container.nginx: Refreshing state... [id=6c830b003efd1939725e78490013225267a625a429780d1c1b918e6ac3cf7650]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # docker_container.nginx must be replaced
-/+ resource "docker_container" "nginx" {
      + bridge                                      = (known after apply)
      ~ command                                     = [
          - "nginx",
          - "-g",
          - "daemon off;",
        ] -> (known after apply)
      + container_logs                              = (known after apply)
      - cpu_shares                                  = 0 -> null
      - dns                                         = [] -> null
      - dns_opts                                    = [] -> null
      - dns_search                                  = [] -> null
      ~ entrypoint                                  = [
          - "/docker-entrypoint.sh",
        ] -> (known after apply)
      ~ env                                         = [] -> (known after apply)
      + exit_code                                   = (known after apply)
      - group_add                                   = [] -> null
      ~ hostname                                    = "6c830b003efd" -> (known after apply)
      ~ id                                          = "6c830b003efd1939725e78490013225267a625a429780d1c1b918e6ac3cf7650" -> (known after apply)
      ~ image                                       = "sha256:247f7abff9f7097bbdab57df76fedd124d1e24a6ec4944fb5ef0ad128997ce05" # forces replacement -> (known after apply) # forces replacement
      ~ init                                        = false -> (known after apply)
      ~ ipc_mode                                    = "private" -> (known after apply)
      ~ log_driver                                  = "json-file" -> (known after apply)
      - log_opts                                    = {} -> null
      - max_retry_count                             = 0 -> null
      - memory                                      = 0 -> null
      - memory_swap                                 = 0 -> null
        name                                        = "tutorial"
      ~ network_data                                = [
          - {
              - gateway                   = "172.17.0.1"
              - global_ipv6_address       = ""
              - global_ipv6_prefix_length = 0
              - ip_address                = "172.17.0.2"
              - ip_prefix_length          = 16
              - ipv6_gateway              = ""
              - mac_address               = "02:42:ac:11:00:02"
              - network_name              = "bridge"
            },
        ] -> (known after apply)
      - network_mode                                = "default" -> null
      - privileged                                  = false -> null
      - publish_all_ports                           = false -> null
      ~ runtime                                     = "runc" -> (known after apply)
      ~ security_opts                               = [] -> (known after apply)
      ~ shm_size                                    = 64 -> (known after apply)
      ~ stop_signal                                 = "SIGQUIT" -> (known after apply)
      ~ stop_timeout                                = 0 -> (known after apply)
      - storage_opts                                = {} -> null
      - sysctls                                     = {} -> null
      - tmpfs                                       = {} -> null
        # (13 unchanged attributes hidden)

      ~ ports {
          ~ external = 8000 -> 8001 # forces replacement
            # (3 unchanged attributes hidden)
        }
    }

  # docker_image.nginx must be replaced
-/+ resource "docker_image" "nginx" {
      ~ id          = "sha256:247f7abff9f7097bbdab57df76fedd124d1e24a6ec4944fb5ef0ad128997ce05nginx:latest" -> (known after apply)
      ~ image_id    = "sha256:247f7abff9f7097bbdab57df76fedd124d1e24a6ec4944fb5ef0ad128997ce05" -> (known after apply)
      ~ name        = "nginx:latest" -> "nginx:alpine" # forces replacement
      ~ repo_digest = "nginx@sha256:ea97e6aace270d82c73da382ea1a8c42d44b9dc11b55159104e21c49c687e7fb" -> (known after apply)
    }

Plan: 2 to add, 0 to change, 2 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

docker_container.nginx: Destroying... [id=6c830b003efd1939725e78490013225267a625a429780d1c1b918e6ac3cf7650]
docker_container.nginx: Destruction complete after 0s
docker_image.nginx: Destroying... [id=sha256:247f7abff9f7097bbdab57df76fedd124d1e24a6ec4944fb5ef0ad128997ce05nginx:latest]
docker_image.nginx: Destruction complete after 1s
docker_image.nginx: Creating...
docker_image.nginx: Creation complete after 5s [id=sha256:2b70e4aaac6b5370bf3a556f5e13156692351696dd5d7c5530d117aa21772748nginx:alpine]
docker_container.nginx: Creating...
docker_container.nginx: Creation complete after 0s [id=c9346b76c28117d69383fbaf27523b1185def2dde87b823fe299333aee38469c]

Apply complete! Resources: 2 added, 0 changed, 2 destroyed.
```

Как видно, оба ресурса были пересозданы:
```console
$ docker images
REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        alpine    2b70e4aaac6b   3 months ago   42.6MB
$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED              STATUS              PORTS                  NAMES
c9346b76c281   2b70e4aaac6b   "/docker-entrypoint.…"   About a minute ago   Up About a minute   0.0.0.0:8001->80/tcp   tutorial
$ curl -I localhost:8001
HTTP/1.1 200 OK
```

## Destroy
Для удаления всех инфраструктурных ресурсов есть команда
`terraform destroy`:
```console
$ terraform destroy
docker_image.nginx: Refreshing state... [id=sha256:2b70e4aaac6b5370bf3a556f5e13156692351696dd5d7c5530d117aa21772748nginx:alpine]
docker_container.nginx: Refreshing state... [id=c9346b76c28117d69383fbaf27523b1185def2dde87b823fe299333aee38469c]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # docker_container.nginx will be destroyed
  - resource "docker_container" "nginx" {
      - attach                                      = false -> null
      - command                                     = [
          - "nginx",
          - "-g",
          - "daemon off;",
        ] -> null
      - container_read_refresh_timeout_milliseconds = 15000 -> null
      - cpu_shares                                  = 0 -> null
      - dns                                         = [] -> null
      - dns_opts                                    = [] -> null
      - dns_search                                  = [] -> null
      - entrypoint                                  = [
          - "/docker-entrypoint.sh",
        ] -> null
      - env                                         = [] -> null
      - group_add                                   = [] -> null
      - hostname                                    = "c9346b76c281" -> null
      - id                                          = "c9346b76c28117d69383fbaf27523b1185def2dde87b823fe299333aee38469c" -> null
      - image                                       = "sha256:2b70e4aaac6b5370bf3a556f5e13156692351696dd5d7c5530d117aa21772748" -> null
      - init                                        = false -> null
      - ipc_mode                                    = "private" -> null
      - log_driver                                  = "json-file" -> null
      - log_opts                                    = {} -> null
      - logs                                        = false -> null
      - max_retry_count                             = 0 -> null
      - memory                                      = 0 -> null
      - memory_swap                                 = 0 -> null
      - must_run                                    = true -> null
      - name                                        = "tutorial" -> null
      - network_data                                = [
          - {
              - gateway                   = "172.17.0.1"
              - global_ipv6_address       = ""
              - global_ipv6_prefix_length = 0
              - ip_address                = "172.17.0.2"
              - ip_prefix_length          = 16
              - ipv6_gateway              = ""
              - mac_address               = "02:42:ac:11:00:02"
              - network_name              = "bridge"
            },
        ] -> null
      - network_mode                                = "default" -> null
      - privileged                                  = false -> null
      - publish_all_ports                           = false -> null
      - read_only                                   = false -> null
      - remove_volumes                              = true -> null
      - restart                                     = "no" -> null
      - rm                                          = false -> null
      - runtime                                     = "runc" -> null
      - security_opts                               = [] -> null
      - shm_size                                    = 64 -> null
      - start                                       = true -> null
      - stdin_open                                  = false -> null
      - stop_signal                                 = "SIGQUIT" -> null
      - stop_timeout                                = 0 -> null
      - storage_opts                                = {} -> null
      - sysctls                                     = {} -> null
      - tmpfs                                       = {} -> null
      - tty                                         = false -> null
      - wait                                        = false -> null
      - wait_timeout                                = 60 -> null

      - ports {
          - external = 8001 -> null
          - internal = 80 -> null
          - ip       = "0.0.0.0" -> null
          - protocol = "tcp" -> null
        }
    }

  # docker_image.nginx will be destroyed
  - resource "docker_image" "nginx" {
      - id          = "sha256:2b70e4aaac6b5370bf3a556f5e13156692351696dd5d7c5530d117aa21772748nginx:alpine" -> null
      - image_id    = "sha256:2b70e4aaac6b5370bf3a556f5e13156692351696dd5d7c5530d117aa21772748" -> null
      - name        = "nginx:alpine" -> null
      - repo_digest = "nginx@sha256:f2802c2a9d09c7aa3ace27445dfc5656ff24355da28e7b958074a0111e3fc076" -> null
    }

Plan: 0 to add, 0 to change, 2 to destroy.

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

docker_container.nginx: Destroying... [id=c9346b76c28117d69383fbaf27523b1185def2dde87b823fe299333aee38469c]
docker_container.nginx: Destruction complete after 1s
docker_image.nginx: Destroying... [id=sha256:2b70e4aaac6b5370bf3a556f5e13156692351696dd5d7c5530d117aa21772748nginx:alpine]
docker_image.nginx: Destruction complete after 0s

Destroy complete! Resources: 2 destroyed.
```
После выполнения - все ресурсы, управляемые [terraform][] будут удалены:
```console
$ docker images
REPOSITORY   TAG       IMAGE ID   CREATED   SIZE
$ docker ps
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
```

[terraform]:https://developer.hashicorp.com/terraform
[docker-provider]:https://github.com/kreuzwerker/terraform-provider-docker/
[language]:https://developer.hashicorp.com/terraform/language#about-the-terraform-language
