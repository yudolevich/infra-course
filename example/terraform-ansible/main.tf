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

locals {
  p2_ip = data.external.p2_host.result.ip
  front_port = 8888
  back_port = 8889
  db_port = 5432
}

variable "db_user" {
  type = string
  validation {
    condition = length(var.db_user) > 0
    error_message = "db_user must be not empty"
  }
}

variable "db_password" {
  type = string
  sensitive = true
  validation {
    condition = length(var.db_password) > 0
    error_message = "db_password must be not empty"
  }
}

provider "docker" {
  alias = "p1"
}

provider "docker" {
  host = "tcp://${local.p2_ip}:2375"
  alias = "p2"
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
