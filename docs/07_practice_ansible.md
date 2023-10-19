# Ansible
В данном практическом занятии рассматривается базовое использование ansible.

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
Скопируем данный файл в директорию проекта с `Vagrantfile` с именем `key`. Все дальнейшие
команды будем вводить находясь на машине `bastion` в директории `/vagrant`.

## Inventory
Для управления другими машинами нам необходимо создать [inventory][] файл с их списком,
который также можно разбить по группам. Создадим файл `hosts`:
```ini
[bastion]
bastion.local

[nodes]
node1.local
node2.local
```

## Basic usage
Управлять машинами из [inventory][] можно из командной строки с помощью [команды `ansible`][cli].
С помощью опции `-i` указывается путь до [inventory][], `--key-file` указывает на `ssh` ключ
с которым будет производиться подключение, `-m` задает [модуль][module] для исполнения и позиционный
аргумент (в данном случае `all`) выбирает на каких машинах будет происходить выполнение.
Для отключения проверки `ssh` ключей при подключении можно выставить переменную среды
`ANSIBLE_HOST_KEY_CHECKING` в значение `False`. Воспользуемся [модулем][module] `ping` и
[запустим проверку всех машин][adhoc]:
```console
$ export ANSIBLE_HOST_KEY_CHECKING=False
$ ansible -i hosts --key-file key -m ping all
bastion.local | SUCCESS => {
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
node2.local | SUCCESS => {
    "ansible_facts": {
        "discovered_interpreter_python": "/usr/bin/python3"
    },
    "changed": false,
    "ping": "pong"
}
```

Можно также выбрать группу для запуска, если не требуется запуск на всех машинах. Воспользуемся
[модулем][module] `shell` для запуска команды на группе `nodes`:
```console
$ ansible -i hosts --key-file key -m shell -a 'uname -a' nodes
node1.local | CHANGED | rc=0 >>
Linux node1 6.2.0-31-generic #31-Ubuntu SMP PREEMPT_DYNAMIC Mon Aug 14 13:42:26 UTC 2023 x86_64 x86_64 x86_64 GNU/Linux
node2.local | CHANGED | rc=0 >>
Linux node2 6.2.0-31-generic #31-Ubuntu SMP PREEMPT_DYNAMIC Mon Aug 14 13:42:26 UTC 2023 x86_64 x86_64 x86_64 GNU/Linux
```

Также ограничить выполнения только определенными хостами/группами можно опцией `-l|--limit`:
```console
$ ansible -i hosts --key-file key -m shell -a 'hostname' nodes -l node1.local
node1.local | CHANGED | rc=0 >>
node1
```

```{note}
Для получения информации по параметрам [модуля][module] можно воспользоваться командой `ansible-doc`,
например `ansible-doc shell`. Получить список [модулей][module] можно командой `ansible-doc -l`.
```

## Playbook
### Tasks
Для декларативного описания задач(tasks), которые будут выполняться на машинах в [inventory][],
можно описать [playbook][] в виде `yaml` файла. Плейбук может состоять из одного или нескольких `play`,
которые в свою очередь содержат задачи(tasks). Также в `play` указывается группа хостов(hosts)
для запуска. Сделаем плейбук, который будет устанавливать пакет `nginx` на группу `nodes` с
помощью [модуля][module] `ansible.builtin.package`:
```yaml
- hosts: nodes
  become: true
  tasks:
  - name: nginx package
    ansible.builtin.package:
      name: nginx
```
Здесь также добавлен параметр `become: true` для запроса повышения привилегий, так как для установки
пакетов в систему требуются права `root`. Сохраним в файл `playbook.yaml` и запустим с помощью
команды `ansible-playbook`:
```console
$ ansible-playbook -i hosts --key-file key playbook.yaml

PLAY [nodes] *******************************************************************************************

TASK [Gathering Facts] *********************************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [nginx package] ***********************************************************************************
changed: [node1.local]
changed: [node2.local]

PLAY RECAP *********************************************************************************************
node1.local                : ok=2    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=2    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
$ curl -s node1.local | grep title
<title>Welcome to nginx!</title>
$ curl -s node2.local | grep title
<title>Welcome to nginx!</title>
```
Как видно `ansible` запускает `play` и в нем запускает ряд задач `tasks`. Первым запускается
задача `Gathering Facts`, которая собирает информацию о машинах во внутренние переменные, которые
потом можно использовать в других задачах. Если информация не нужна, то можно добавить параметр
`gather_facts: False`:
```yaml
---
- hosts: nodes
  become: True
  gather_facts: False
  tasks:
  - name: nginx package
    ansible.builtin.package:
      name: nginx
```
```console
$ ansible-playbook -i hosts --key-file key playbook.yaml

PLAY [nodes] *******************************************************************************************

TASK [nginx package] ***********************************************************************************
ok: [node1.local]
ok: [node2.local]

PLAY RECAP *********************************************************************************************
node1.local                : ok=1    changed=0    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=1    changed=0    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
```
Как видно задача по сбору фактов пропала, а также видно что задача `nginx package` теперь находится
не в статусе `changed`, а в статусе `ok`, что означает что никаких действий не было произведено, так
как пакет был установлен при прошлом запуске.

### Handlers
Для того чтобы [запускать какие-либо действия только при изменениях][handlers] в `play` также
можно указать список `handlers`, задачи в котором будут выполняться только при наступлении
определенных событий.
Добавим в основной раздел `tasks` задачу по копированию конфигурации, а в `handlers` рестарт
сервиса `nginx`, который будет запускаться при изменении конфигурации:
```yaml
- hosts: nodes
  become: True
  gather_facts: False
  tasks:
  - name: nginx package
    ansible.builtin.package:
      name: nginx
  - name: nginx config
    ansible.builtin.copy:
      src: default
      dest: /etc/nginx/sites-enabled/default
    notify: nginx restart

  handlers:
  - name: nginx restart
    ansible.builtin.systemd:
      name: nginx
      state: restarted
```
Сам файл конфигурации расположим рядом с плейбуком в директории `files`, которая используется
по умолчанию для копирования файлов на удаленные машины. Файл конфигурации будет называться
`default`:
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

```console
$ ansible-playbook -i hosts --key-file key playbook.yaml

PLAY [nodes] *****************************************************************************

TASK [nginx package] *********************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [nginx config] **********************************************************************
changed: [node1.local]
changed: [node2.local]

RUNNING HANDLER [nginx restart] **********************************************************
changed: [node1.local]
changed: [node2.local]

PLAY RECAP *******************************************************************************
node1.local                : ok=3    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=3    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
```

### Jinja2
В `ansible` [используются шаблонизатор `jinja2`][jinja] для использования переменных как в самих
плейбуках, так и в [модуле][module] `template` для шаблонизации файлов на удаленных машинах.
Создадим в директории `templates` шаблон файл `index.html.j2`:
```jinja
hello from {{ ansible_host }}
```
Где `ansible_host` будет указывать на машину, на которой будет выполняться задача. Дополним плейбук:
```yaml
- hosts: nodes
  become: True
  gather_facts: False
  vars:
    html_dir: /var/www/html
  tasks:
  - name: nginx package
    ansible.builtin.package:
      name: nginx
  - name: nginx config
    ansible.builtin.copy:
      src: default
      dest: /etc/nginx/sites-enabled/default
    notify: nginx restart
  - name: index.html to "{{ html_dir }}"
    ansible.builtin.template:
      src: index.html.j2
      dest: "{{ html_dir }}/index.html"

  handlers:
  - name: nginx restart
    ansible.builtin.systemd:
      name: nginx
      state: restarted
```
Здесь мы также вынесли в блок с переменными `vars` путь до html директории, чтобы показать как
можно использовать переменные в самом плейбуке.

```console
$ ansible-playbook -i hosts --key-file key playbook.yaml

PLAY [nodes] ******************************************************************************************

TASK [nginx package] **********************************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [nginx config] ***********************************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [index.html to "/var/www/html"] ******************************************************************
changed: [node1.local]
changed: [node2.local]

PLAY RECAP ********************************************************************************************
node1.local                : ok=3    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=3    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0

$ curl node1.local
hello from node1.local
$ curl node2.local
hello from node2.local
```

### Conditionals
Для задачи [можно определить условия, при которых она должна выполняться][conditional] с помощью
директивы `when`:
```yaml
---
- hosts: nodes
  become: True
  gather_facts: False
  vars:
    html_dir: /var/www/html
  tasks:
  - name: nginx package
    ansible.builtin.package:
      name: nginx
  - name: nginx config
    ansible.builtin.copy:
      src: default
      dest: /etc/nginx/sites-enabled/default
    notify: nginx restart
  - name: index.html to "{{ html_dir }}"
    ansible.builtin.template:
      src: index.html.j2
      dest: "{{ html_dir }}/index.html"
  - name: node1only file
    ansible.builtin.copy:
      content: "hello from node1 only\n"
      dest: "{{ html_dir }}/node1only"
    when: ansible_host == "node1.local"

  handlers:
  - name: nginx restart
    ansible.builtin.systemd:
      name: nginx
      state: restarted
```

```console
$ ansible-playbook -i hosts --key-file key playbook.yaml

PLAY [nodes] ******************************************************************************************

TASK [nginx package] **********************************************************************************
ok: [node2.local]
ok: [node1.local]

TASK [nginx config] ***********************************************************************************
ok: [node2.local]
ok: [node1.local]

TASK [index.html to "/var/www/html"] ******************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [node1only file] *********************************************************************************
skipping: [node2.local]
changed: [node1.local]

PLAY RECAP ********************************************************************************************
node1.local                : ok=4    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
$ curl node1.local/node1only
hello from node1 only
$ curl node2.local/node1only
<html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>nginx/1.22.0 (Ubuntu)</center>
</body>
</html>
```
Как видно задача на машине `node2.local` не выполнялась, так как для нее не выполнено условие в
директиве `when`.

### Loops
В `ansible` есть [несколько директив для определения цикла][loops]: `loop`, `with_*`, `until`.
С помощью них можно выполнить задачу множество раз с разными переменными.
Добавим задачу с циклом в плейбук:
```yaml
---
- hosts: nodes
  become: True
  gather_facts: False
  vars:
    html_dir: /var/www/html
  tasks:
  - name: nginx package
    ansible.builtin.package:
      name: nginx
  - name: nginx config
    ansible.builtin.copy:
      src: default
      dest: /etc/nginx/sites-enabled/default
    notify: nginx restart
  - name: index.html to "{{ html_dir }}"
    ansible.builtin.template:
      src: index.html.j2
      dest: "{{ html_dir }}/index.html"
  - name: node1only file
    ansible.builtin.copy:
      content: "hello from node1 only\n"
      dest: "{{ html_dir }}/node1only"
    when: ansible_host == "node1.local"
  - name: loop files
    ansible.builtin.copy:
      content: "{{ ansible_host }}\n"
      dest: "{{ html_dir}}/{{ item }}"
    loop:
      - test1
      - test2
      - test3

  handlers:
  - name: nginx restart
    ansible.builtin.systemd:
      name: nginx
      state: restarted
```

```console
$ ansible-playbook -i hosts --key-file key playbook.yaml

PLAY [nodes] ******************************************************************************************

TASK [nginx package] **********************************************************************************
ok: [node2.local]
ok: [node1.local]

TASK [nginx config] ***********************************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [index.html to "/var/www/html"] ******************************************************************
ok: [node1.local]
ok: [node2.local]

TASK [node1only file] *********************************************************************************
skipping: [node2.local]
ok: [node1.local]

TASK [loop files] *************************************************************************************
changed: [node1.local] => (item=test1)
changed: [node2.local] => (item=test1)
changed: [node1.local] => (item=test2)
changed: [node2.local] => (item=test2)
changed: [node1.local] => (item=test3)
changed: [node2.local] => (item=test3)

PLAY RECAP ********************************************************************************************
node1.local                : ok=5    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
node2.local                : ok=4    changed=1    unreachable=0    failed=0    skipped=1    rescued=0    ignored=0

$ curl node1.local/test1
node1.local
$ curl node2.local/test2
node2.local
$ curl node1.local/test3
node1.local
```

[inventory]:https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html
[cli]:https://docs.ansible.com/ansible/latest/cli/ansible.html
[adhoc]:https://docs.ansible.com/ansible/latest/command_guide/intro_adhoc.html
[module]:https://docs.ansible.com/ansible/latest/module_plugin_guide/modules_intro.html
[playbook]:https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_intro.html
[handlers]:https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_handlers.html
[jinja]:https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_templating.html
[conditional]:https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_conditionals.html
[loops]:https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_loops.html
