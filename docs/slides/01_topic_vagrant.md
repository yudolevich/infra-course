## Vagrant

```{image} ../img/vagrant.svg
:width: 200px
```

### Основные концепции

```{revealjs-fragments}
* Vagrantfile
* Boxes
* Providers
* Provisioning
```

### Vagrantfile
```{revealjs-code-block} ruby
Vagrant.configure("2") do |config|
  # ...
end
```

### Vagrantfile lookup
```{revealjs-code-block} bash
---
data-line-numbers: 1-5|7
---
/home/mitchellh/projects/foo/Vagrantfile
/home/mitchellh/projects/Vagrantfile
/home/mitchellh/Vagrantfile
/home/Vagrantfile
/Vagrantfile

$VAGRANT_CWD
```

### Vagrantfile merging
```{revealjs-fragments}
* Box
* Home (~/.vagrant.d)
* Project
* Multi-machine
* Provider
```

### Box
```{revealjs-code-block} console
$ vagrant init hashicorp/bionic64
```
```{revealjs-code-block} ruby
Vagrant.configure("2") do |config|
  config.vm.box = "hashicorp/bionic64"
end
```

### Box
```{revealjs-code-block} console
---
data-line-numbers: 1|2|3-4|5|6|7-8
---
$ ls ~/.vagrant.d/boxes/ubuntu-VAGRANTSLASH-lunar64/0/virtualbox
box.ovf
metadata.json
ubuntu-lunar-23.04-cloudimg-configdrive.vmdk
ubuntu-lunar-23.04-cloudimg.vmdk
ubuntu-lunar-23.04-cloudimg.mf
Vagrantfile
vagrant_insecure_key
vagrant_insecure_key.pub
```

### Provider
```{revealjs-fragments}
* VirtualBox
* VMware
* Hyper-V
* Libvirt
* Docker
```

### Provider configuration
```{revealjs-code-block} ruby
Vagrant.configure("2") do |config|
  # ...
  config.vm.provider "virtualbox" do |vb|
    vb.customize ["modifyvm", :id, "--cpuexecutioncap", "50"]
    vb.memory = 1024
    vb.cpus = 2
  end
end
```

### Provider Usage
```{revealjs-code-block} console
---
data-line-numbers: 1-3|4
---
$ vagrant box list
bionic64 (virtualbox)
bionic64 (vmware_fusion)
$ vagrant up --provider=vmware_fusion
```

### Provision
```{revealjs-code-block} ruby
Vagrant.configure("2") do |config|
  # ... other configuration
  config.vm.provision "shell" do |s|
    s.inline = "echo hello"
  end
end
```

### Provision run
```{revealjs-code-block} console
$ vagrant up
$ vagrant reload --provision
$ vagrant provision
```

### Provision run
```{revealjs-code-block} ruby
Vagrant.configure("2") do |config|
  # ... other configuration
  config.vm.provision "bootstrap", type: "shell" do |s|
    s.inline = "echo hello"
  end
end
```

```{revealjs-code-block} console
$ vagrant provision --provision-with bootstrap
```

### Provision run
```{revealjs-code-block} ruby
Vagrant.configure("2") do |config|
  config.vm.provision "bootstrap", type: "shell", run: "never" do |s|
    s.inline = "echo hello"
  end
end
```

### Provision type
```{revealjs-fragments}
* File
* Shell
* Ansible
* Docker
* Chef
* Puppet
* Salt
```

### Networking
```{revealjs-code-block} ruby
---
data-line-numbers: 3|4|5|6
---
Vagrant.configure("2") do |config|
  # ...
  config.vm.network "forwarded_port", guest: 80, host: 8080
  config.vm.hostname = "myhost.local"
  config.vm.network "private_network", ip: "192.168.50.4"
  config.vm.network "public_network", ip: "192.168.0.17"
end
```

### Synced Folders
```{revealjs-code-block} ruby
---
data-line-numbers: 3|4
---
Vagrant.configure("2") do |config|
  # other config here
  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.vm.synced_folder "src/", "/srv/website"
end
```

### Synced Folders Types
```{revealjs-code-block} ruby
---
data-line-numbers: 2|3|4
---
Vagrant.configure("2") do |config|
  config.vm.synced_folder ".", "/vagrant", type: "nfs"
  config.vm.synced_folder ".", "/vagrant", type: "rsync"
  config.vm.synced_folder ".", "/vagrant", type: "smb"
end
```

### Plugins
```{revealjs-code-block} console
## Installing a plugin from a known gem source
$ vagrant plugin install my-plugin
## Installing a plugin from a local file source
$ vagrant plugin install /path/to/my-plugin.gem
```
