# Ansible Roles and Collections
Данное практическое занятие посвящено знакомству с ролями(roles)
и коллекциями(collections) в ansible.

## Vagrant
Для работы с ansible воспользуемся следующим `Vagrantfile` c тремя машинами:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "bastion" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "bastion"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns ansible
    SHELL
  end

  config.vm.define "node1" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "node1"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns
    SHELL
  end

  config.vm.define "node2" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "node2"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns
    SHELL
  end

end
```
Все управление будем производить с машины `bastion`, для взаимодействия с другими машинами
потребуется `ssh` ключ, путь до которого можно узнать в выводе команды `vagrant ssh-config`:
```console
$ vagrant ssh-config bastion | grep IdentityFile
  IdentityFile ~/.vagrant.d/boxes/ubuntu-VAGRANTSLASH-lunar64/0/virtualbox/vagrant_insecure_key
```
Скопируем данный файл в директорию проекта с `Vagrantfile` с именем `key`.
Все дальнейшие команды будем вводить находясь на машине `bastion`, в первую очередь добавив
ключ пользователю `vagrant` и выставив переменную `ANSIBLE_HOST_KEY_CHECKING` в значение
`False` для отключения проверки `ssh` ключей.
```console
$ install -m 600 -o vagrant /vagrant/key /home/vagrant/.ssh/id_rsa
$ export ANSIBLE_HOST_KEY_CHECKING=False
```

В качестве `inventory` можно использовать список хостов через запятую:
```console
$ ansible -m ping -i node1.local,node2.local all
node2.local | SUCCESS => {
    "ansible_facts": {
        "discovered_interpreter_python": "/usr/bin/python3"
    },
    "changed": false,
    "ping": "pong"
}
node1.local | SUCCESS => {
    "ansible_facts": {
        "discovered_interpreter_python": "/usr/bin/python3"
    },
    "changed": false,
    "ping": "pong"
}
```

## Roles
Для того чтобы не писать однотипные плейбуки для часто повторяемых операций,
а также для логического разбиения и переиспользования в `ansible` существует
[концепция ролей][roles].

### Init
[Роль][roles] имеет определенную структуру директорий,
чтобы создать данную структуру можно воспользоваться командой `ansible-galaxy`:
```console
$ ansible-galaxy role init --init-path roles nginx
$ ls -1 roles/nginx/
README.md  # описание роли в формате markdown
defaults   # переменные по-умолчанию с низким приоритетом
files      # файлы используемые в задачах, например для копирования
handlers   # хендлеры, выполняемы по событиям
meta       # мета информация о роли, например об авторе и лицензии
tasks      # основные задачи выполняемые ролью
templates  # jinja2 шаблоны используемые в задачах
tests      # тесты для проверки роли
vars       # переменные для этой роли
```
После выполнения команды создалась структура каталогов для роли с именем `nginx`.

### Tasks
Добавим в роль установку пакета, для это допишем в файл `roles/nginx/tasks/main.yml`:
```yaml
---
# tasks file for nginx
- name: nginx package
  ansible.builtin.package:
    name: nginx
```

[Роли][roles] можно использовать в плейбуках также как задачи(tasks) с помощью
директив `roles`, `include_role` и `import_role`. Создадим файл `playbook.yaml`:
```yaml
---
- hosts: all
  become: True
  roles:
  - name: nginx
```

И запустим:
```console
$ ansible-playbook -i node1.local,node2.local playbook.yaml

PLAY [all] ********************************************************************************

TASK [Gathering Facts] ********************************************************************
ok: [node2.local]
ok: [node1.local]

TASK [nginx : nginx package] **************************************************************
changed: [node1.local]
changed: [node2.local]

PLAY RECAP ********************************************************************************
node1.local                : ok=2    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=2    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
$ curl -s node1.local | grep title
<title>Welcome to nginx!</title>
```

### Defaults
Для задания значений по-умолчанию к переменным используемым в роли можно отредактировать
файл `roles/nginx/defaults/main.yml`:
```yaml
---
# defaults file for nginx
nginx_html_dir: /var/www/html
nginx_config_file: default
nginx_template_file: index.html.j2
```

А также добавим соответствующие файлы
`roles/nginx/files/default`:
```nginx
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    location / {
        root   /var/www/html;
        index  index.html;
    }
}
```
`roles/nginx/templates/index.html.j2`:
```jinja
hello from {{ ansible_host }}
```

### Files/Templates
Добавим задачи по копированию конфигурации и `index.html` в нашу роль
в файл `roles/nginx/tasks/main.yml`:
```yaml
---
# tasks file for nginx
- name: nginx package
  ansible.builtin.package:
    name: nginx
- name: nginx config
  ansible.builtin.copy:
    src: "{{ nginx_config_file }}"
    dest: "/etc/nginx/sites-enabled/default"
- name: "index.html to {{ nginx_html_dir }}"
  ansible.builtin.template:
    src: "{{ nginx_template_file }}"
    dest: "{{ nginx_html_dir }}/index.html"
```

И запустим плейбук:
```console
$ ansible-playbook -i node1.local,node2.local playbook.yaml

PLAY [all] ********************************************************************************

TASK [Gathering Facts] ********************************************************************
ok: [node2.local]
ok: [node1.local]

TASK [nginx : nginx package] **************************************************************
ok: [node2.local]
ok: [node1.local]

TASK [nginx : nginx config] ***************************************************************
changed: [node1.local]
changed: [node2.local]

TASK [nginx : index.html to /var/www/html] ************************************************
changed: [node1.local]
changed: [node2.local]

PLAY RECAP ********************************************************************************
node1.local                : ok=4    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=4    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
$ curl -s node1.local
hello from node1.local
$ curl -s node2.local
hello from node2.local
```

Для переопределения значений по-умолчанию расположим в директориях `files` и `templates`
рядом с плейбуком
`templates/overrided.html.j2`:
```jinja
### hello from {{ ansible_host }} !
```
`files/overrided.conf`:
```nginx
server {
    listen       80;
    listen  [::]:80;
    listen     8080;
    server_name  localhost;

    location / {
        root   /var/www/html;
        index  index.html;
    }
}
```

И добавим переменные в плейбук:
```yaml
---
- hosts: all
  become: True
  roles:
  - name: nginx
    nginx_template_file: overrided.html.j2
    nginx_config_file: overrided.conf
```

Запустим:
```console
$ ansible-playbook -i node1.local,node2.local playbook.yaml

PLAY [all] ********************************************************************************

TASK [Gathering Facts] ********************************************************************
ok: [node2.local]
ok: [node1.local]

TASK [nginx : nginx package] **************************************************************
ok: [node2.local]
ok: [node1.local]

TASK [nginx : nginx config] ***************************************************************
changed: [node2.local]
changed: [node1.local]

TASK [nginx : index.html to /var/www/html] ************************************************
changed: [node2.local]
changed: [node1.local]

PLAY RECAP ********************************************************************************
node1.local                : ok=4    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=4    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
$ curl node1.local
### hello from node1.local !
$ curl node2.local
### hello from node2.local !
$ curl node1.local:8080
curl: (7) Failed to connect to node1.local port 8080 after 3 ms: Couldn't connect to server
```
Как видно мы переопределили значения из роли не меняя саму роль. Но измененная
конфигурация `nginx` хоть и скопировалась, но не применилась.

### Handlers
Для перезапуска `nginx` при изменении конфигурации добавим в файл
`roles/nginx/handlers/main.yml`:
```yaml
---
# handlers file for nginx
- name: nginx restart
  ansible.builtin.systemd:
    name: nginx
    state: restarted
```
И соответственно в задачу по записи конфигурации отправку события в файле
`roles/nginx/tasks/main.yml`:
```yaml
---
# tasks file for nginx
- name: nginx package
  ansible.builtin.package:
    name: nginx
- name: nginx config
  ansible.builtin.copy:
    src: "{{ nginx_config_file }}"
    dest: "/etc/nginx/sites-enabled/default"
  notify: nginx restart
- name: "index.html to {{ nginx_html_dir }}"
  ansible.builtin.template:
    src: "{{ nginx_template_file }}"
    dest: "{{ nginx_html_dir }}/index.html"
```

Теперь можем изменить нашу конфигурацию в файле `files/overrided.conf`:
```nginx
server {
    listen       80;
    listen     8080;
    server_name  localhost;

    location / {
        root   /var/www/html;
        index  index.html;
    }
}
```

И запустить плейбук:
```yaml
$ ansible-playbook -i node1.local,node2.local playbook.yaml

PLAY [all] ********************************************************************************

TASK [Gathering Facts] ********************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [nginx : nginx package] **************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [nginx : nginx config] ***************************************************************
changed: [node1.local]
changed: [node2.local]

TASK [nginx : index.html to /var/www/html] ************************************************
ok: [node1.local]
ok: [node2.local]

RUNNING HANDLER [nginx : nginx restart] ***************************************************
changed: [node2.local]
changed: [node1.local]

PLAY RECAP ********************************************************************************
node1.local                : ok=5    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=5    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
$ curl node1.local:8080
### hello from node1.local !
$ curl node2.local:8080
### hello from node2.local !
```

Таким образом мы получили роль, которую можно переиспользовать с другими значениями
переменных.

## Collections
[Коллекции][collections] в `ansible` - это способ создания переносимых дистрибутивов,
которые могут включать в себя плейбуки, роли, модули и плагины.

Перед работой с коллекциями пересоздадим виртуальные машины командами `vagrant destroy -f`
и `vagrant up`. Примеры для работы с коллекциями также будут созданы с нуля в пустой
директории.

### Install
[Коллекциями][collections] как и [ролями][roles] можно управлять с помощью утилиты
[`ansible-galaxy`][galaxy-cli]. Данная утилита по-умолчанию использует публичный
репозиторий [galaxy.ansible.com](https://galaxy.ansible.com), в котором можно найти
готовые коллекции от комьюнити. Установим из данного репозитория коллекцию
[nginxinx.nginx_core][nginxcore]:
```console
$ ansible-galaxy collection install nginxinc.nginx_core
Starting galaxy collection install process
Process install dependency map
Starting collection install process
Downloading https://galaxy.ansible.com/api/v3/plugin/ansible/content/published/collections/artifacts/nginxinc-nginx_core-0.8.0.tar.gz to /home/vagrant/.ansible/tmp/ansible-local-31197c0gyeuh/tmpfiqc995o/nginxinc-nginx_core-0.8.0-o05zpo12
Installing 'nginxinc.nginx_core:0.8.0' to '/home/vagrant/.ansible/collections/ansible_collections/nginxinc/nginx_core'
Downloading https://galaxy.ansible.com/api/v3/plugin/ansible/content/published/collections/artifacts/ansible-posix-1.5.4.tar.gz to /home/vagrant/.ansible/tmp/ansible-local-31197c0gyeuh/tmpfiqc995o/ansible-posix-1.5.4-wvjd1tyo
nginxinc.nginx_core:0.8.0 was installed successfully
Installing 'ansible.posix:1.5.4' to '/home/vagrant/.ansible/collections/ansible_collections/ansible/posix'
Downloading https://galaxy.ansible.com/api/v3/plugin/ansible/content/published/collections/artifacts/community-crypto-2.15.1.tar.gz to /home/vagrant/.ansible/tmp/ansible-local-31197c0gyeuh/tmpfiqc995o/community-crypto-2.15.1-h5a72qdu
ansible.posix:1.5.4 was installed successfully
Installing 'community.crypto:2.15.1' to '/home/vagrant/.ansible/collections/ansible_collections/community/crypto'
Downloading https://galaxy.ansible.com/api/v3/plugin/ansible/content/published/collections/artifacts/community-general-7.5.0.tar.gz to /home/vagrant/.ansible/tmp/ansible-local-31197c0gyeuh/tmpfiqc995o/community-general-7.5.0-3w9itazy
community.crypto:2.15.1 was installed successfully
Installing 'community.general:7.5.0' to '/home/vagrant/.ansible/collections/ansible_collections/community/general'
community.general:7.5.0 was installed successfully
```

### Usage
Использовать коллекции в плейбуке можно объявив их в директиве `collections` в `play`,
либо указывая полное имя непосредственно в месте использования:
```yaml
---
- hosts: all
  become: True
  collections:
  - nginxinc.nginx_core
  roles:
  - name: nginxinc.nginx_core.nginx
  - name: nginx_config
    nginx_config_http_template_enable: true
    nginx_config_http_template:
    - config:
        servers:
          - core:
              listen:
                - port: 80
              server_name: localhost
            locations:
              - location: /
                core:
                  root: /usr/share/nginx/html
                  index: index.html
```
Здесь мы используем две роли из коллекции `nginxinc.nginx_core`:
- `nginx` - используя полное имя, данная роль установит сам `nginx`
- `nginx_config` - использую только имя роли, так как имя коллекции было объявлено
  ранее, данная роль позволяет сконфигурировать `nginx`

Запустим получившийся плейбук:
```console
$ ansible-playbook -i node1.local,node2.local playbook.yaml

PLAY [all] ********************************************************************************

TASK [Gathering Facts] ********************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [nginxinc.nginx_core.nginx : Validate distribution and role variables] ***************
...
RUNNING HANDLER [nginxinc.nginx_core.nginx_config : (Handler - NGINX Config) Start/reload NGINX] ***
changed: [node1.local]
changed: [node2.local]

TASK [nginxinc.nginx_core.nginx_config : Debug output] ************************************
skipping: [node1.local]
skipping: [node2.local]

PLAY RECAP ********************************************************************************
node1.local                : ok=22   changed=2    unreachable=0    failed=0    skipped=37   rescued=0    ignored=1
node2.local                : ok=22   changed=2    unreachable=0    failed=0    skipped=37   rescued=0    ignored=1
$ curl -s node1.local | grep title
<title>Welcome to nginx!</title>
$ curl -s node2.local | grep title
<title>Welcome to nginx!</title>
```

Также [ознакомившись с документацией][nginx_config] к роли `nginx_config`, либо же
посмотреть на [список определенных переменных][nginx_config_def], можно узнать что
авторы данной роли позволяют параметризовать. Добавим также шаблон для `index.html`,
как мы это делали в собственной роли в файле `templates/index.html.j2` рядом с плейбуком:
```jinja
hello from {{ ansible_host }}
```

И добавим переменные для роли в плейбук:
```yaml
---
- hosts: all
  become: True
  collections:
  - nginxinc.nginx_core
  roles:
  - name: nginxinc.nginx_core.nginx
  - name: nginx_config
    nginx_config_http_template_enable: true
    nginx_config_http_template:
    - config:
        servers:
        - core:
            listen:
            - port: 80
            server_name: localhost
          locations:
          - location: /
            core:
              root: /usr/share/nginx/html
              index: index.html
    nginx_config_html_demo_template_enable: true
    nginx_config_html_demo_template:
    - template_file: index.html.j2
      deployment_location: /usr/share/nginx/html/index.html
```

Запустим плейбук:
```console
$ ansible-playbook -i node1.local,node2.local playbook.yaml

PLAY [all] ********************************************************************************

TASK [Gathering Facts] ********************************************************************
ok: [node2.local]
ok: [node1.local]
...
PLAY RECAP ********************************************************************************
node1.local                : ok=22   changed=1    unreachable=0    failed=0    skipped=34   rescued=0    ignored=1
node2.local                : ok=22   changed=1    unreachable=0    failed=0    skipped=34   rescued=0    ignored=1

$ curl node1.local
hello from node1.local
$ curl node2.local
hello from node2.local
```

Как видно используя роли и коллекции можно переиспользовать код (в том числе
подготовленный комьюнити и загруженный в [ansible galaxy](https://galaxy.ansible.com))
для автоматизации развертывания инфраструктурных приложений.


[roles]:https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_reuse_roles.html
[collections]:https://docs.ansible.com/ansible/latest/collections_guide/index.html
[galaxy-cli]:https://docs.ansible.com/ansible/latest/collections_guide/collections_installing.html
[nginxcore]:https://galaxy.ansible.com/ui/repo/published/nginxinc/nginx_core/
[nginx_config]:https://galaxy.ansible.com/ui/repo/published/nginxinc/nginx_core/content/role/nginx_config/
[nginx_config_def]:https://github.com/nginxinc/ansible-role-nginx-config/blob/main/defaults/main/template.yml
