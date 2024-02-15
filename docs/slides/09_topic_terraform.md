## Terraform

```{image} ../img/terraform.svg
:width: 200px
```

###
```{image} ../img/terraform1.jpg
:width: 700px
```

###
```{image} ../img/terraform2.jpg
:width: 700px
```

### Concepts
```{revealjs-fragments}
* Configuration
* State
* Provider
```

### Configuration
```{revealjs-code-block} tf
resource "aws_vpc" "main" {
  cidr_block = var.base_cidr_block
}

<BLOCK TYPE> "<BLOCK LABEL>" "<BLOCK LABEL>" {
  # Block body
  <IDENTIFIER> = <EXPRESSION> # Argument
}
```

### Terraform
```{revealjs-code-block} tf
terraform {
  required_providers {
    aws = {
      version = ">= 2.7.0"
      source = "hashicorp/aws"
    }
  }
  experiments = [example]
}
```

### Provider
```{revealjs-code-block} tf
terraform {
  required_providers {
    mycloud = {
      source  = "mycorp/mycloud"
      version = "~> 1.0"
    }
  }
}

provider "mycloud" {
  # ...
}
```

### Resource
```{revealjs-code-block} tf
resource "aws_instance" "web" {
  ami           = "ami-a1b2c3d4"
  instance_type = "t2.micro"
}
```

### Variables
```{revealjs-code-block} tf
---
data-line-numbers: 1-9|10-24|25-28
---
variable "image_id" {
  type = string
}

variable "availability_zone_names" {
  type    = list(string)
  default = ["us-west-1a"]
}

variable "docker_ports" {
  type = list(object({
    internal = number
    external = number
    protocol = string
  }))
  default = [
    {
      internal = 8300
      external = 8300
      protocol = "tcp"
    }
  ]
}

resource "aws_instance" "example" {
  instance_type = "t2.micro"
  ami           = var.image_id
}
```
```{revealjs-code-block} console
$ terraform apply -var="image_id=ami-abc123"
$ terraform apply -var-file="testing.tfvars"
```

### Data Sources
```{revealjs-code-block} tf
data "aws_ami" "example" {
  most_recent = true

  owners = ["self"]
  tags = {
    Name   = "app-server"
    Tested = "true"
  }
}
```
