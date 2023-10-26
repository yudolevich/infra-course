## Ansible

```{image} ../img/ansible-icon.svg
:width: 200px
```

### Basic

```{image} ../img/ansible-basic.svg
:width: 450px
```

### Concepts

```{revealjs-fragments}
* Control/Managed nodes
* Inventory
* Playbooks
* Modules
* Plugins
* Collections
```

### Inventory
```ini
mail.example.com

[webservers]
foo.example.com
bar.example.com

[dbservers]
one.example.com
two.example.com
three.example.com
```

### Inventory
```{revealjs-code-block} yaml
ungrouped:
  hosts:
    mail.example.com:
webservers:
  hosts:
    foo.example.com:
    bar.example.com:
dbservers:
  hosts:
    one.example.com:
    two.example.com:
    three.example.com:
```

### Inventory
```
inventory/
  01-openstack.yml          # configure inventory plugin
  02-dynamic-inventory.py   # dynamic inventory script
  03-static-inventory       # add static hosts
  group_vars/
    all.yml                 # variables to all hosts
```

### Playbook
```console
$ ansible-playbook
$ ansible-pull
$ ansible-lint
```

### Plays and tasks

```{revealjs-code-block} yaml
---
data-line-numbers: 1|16|5-15
---
- name: Update web servers
  hosts: webservers
  remote_user: root

  tasks:
  - name: Ensure apache is at the latest version
    ansible.builtin.yum:
      name: httpd
      state: latest

  - name: Write the apache config file
    ansible.builtin.template:
      src: /srv/httpd.j2
      dest: /etc/httpd.conf

- name: Update db servers
  hosts: databases
  remote_user: root

  tasks:
  - name: Ensure postgresql is at the latest version
    ansible.builtin.yum:
      name: postgresql
      state: latest

  - name: Ensure that postgresql is started
    ansible.builtin.service:
      name: postgresql
      state: started
```
### Handlers

```{revealjs-code-block} yaml
---
data-line-numbers: 1-5|6-8|10-19|21-24|26-37
---
tasks:
- name: Template configuration file
  ansible.builtin.template:
    src: template.j2
    dest: /etc/foo.conf
  notify:
    - Restart apache
    - Restart memcached

handlers:
  - name: Restart memcached
    ansible.builtin.service:
      name: memcached
      state: restarted

  - name: Restart apache
    ansible.builtin.service:
      name: apache
      state: restarted
---
tasks:
  - name: Restart everything
    command: echo "this task will restart the web services"
    notify: "restart web services"

handlers:
  - name: Restart memcached
    service:
      name: memcached
      state: restarted
    listen: "restart web services"

  - name: Restart apache
    service:
      name: apache
      state: restarted
    listen: "restart web services"
```

### Variables

```{revealjs-code-block} yaml
---
data-line-numbers: 1-3|5-10|12-16|18-21
---
- hosts: app_servers
  vars:
      app_path: {{ base_path }}/22
---
- hosts: all
  remote_user: root
  vars:
    favcolor: blue
  vars_files:
    - /vars/external_vars.yml
---
- hosts: web_servers
  tasks:
     - name: Run a shell command
       ansible.builtin.shell: /usr/bin/foo
       register: foo_result
---
- name: Setting host facts using complex arguments
  ansible.builtin.set_fact:
    one_fact: something
    other_fact: "{{ local_var * 2 }}"
```

### Conditionals

```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|6-9
---
tasks:
  - name: Shut down Debian flavored systems
    ansible.builtin.command: /sbin/shutdown -t now
    when: ansible_facts['os_family'] == "Debian"
---
tasks:
  - ansible.builtin.shell: echo "only on Red Hat 6"
    when: ansible_facts['os_family'] == "RedHat" and
          ansible_facts['lsb']['major_release'] | int >= 6
```

### Loops

```{revealjs-code-block} yaml
---
data-line-numbers: 1-8|10-17
---
- name: Add several users
  ansible.builtin.user:
    name: "{{ item }}"
    state: present
    groups: "wheel"
  loop:
     - testuser1
     - testuser2
---
- name: Fail if return code is not 0
  ansible.builtin.fail:
    msg: "The command ({{ it.cmd }}) did not have a 0 return code"
  when: it.rc != 0
  loop: "{{ echo.results }}"
  loop_control:
    pause: 3
    loop_var: it
```

### Tags

```{revealjs-code-block} yaml
tasks:
- name: Install the servers
  ansible.builtin.yum:
    name:
    - httpd
    state: present
  tags:
  - packages
  - webservers
```

```console
$ ansible-playbook --tags packages
```

### Tags

```{revealjs-code-block} yaml
tasks:
- name: Run the rarely-used debug task
  ansible.builtin.debug:
   msg: '{{ showmevar }}'
  tags: [ never, debug ]

- name: Configure the service
  ansible.builtin.template:
    src: templates/src.j2
    dest: /etc/foo.conf
  tags:
  - always
  - template
```

```console
$ ansible-playbook --tags debug --skip-tags template
```

### Jinja2

```yaml
---
- name: Write hostname
  hosts: all
  tasks:
  - name: "write hostname {{ ansible_facts['hostname'] }}"
    ansible.builtin.template:
       src: templates/test.j2
       dest: /tmp/hostname
```

```jinja
My name is {{ ansible_facts['hostname'] }}
{% if True %}
    yay
{% endif %}
{% for i in seq -%}
    {{ i }}
{%- endfor %}
```

### Filters

```jinja
{{ some_variable | default(5) }}
{{ dict | dict2items }}
{{ some_variable | to_json }}
```

### Lookups

```yaml
vars:
  motd_value: "{{ lookup('file', '/etc/motd') }}"
tasks:
  - debug:
      msg: "motd value is {{ motd_value }}"
```

### Roles

```{revealjs-code-block}
---
data-line-numbers: 1-2|3-4|5-6|7-8|9-11|12-13|14-15|16-17|18-20
---
roles/
    common/
        tasks/
            main.yml
        handlers/
            main.yml
        templates/
            ntp.conf.j2
        files/
            bar.txt
            foo.sh
        vars/
            main.yml
        defaults/
            main.yml
        meta/
            main.yml
        library/
        module_utils/
        lookup_plugins/
```

### Roles

```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|6-10
---
- hosts: webservers
  roles:
    - common
    - webservers
---
- hosts: webservers
  tasks:
    - name: Include the foo_app_instance role
      include_role:
        name: foo_app_instance
```

### Modules

```console
$ ansible webservers -m service -a "name=httpd state=started"
$ ansible webservers -m ping
$ ansible webservers -m command -a "/sbin/reboot -t now"
$ ansible-doc command
$ ansible-doc -l
```

```{revealjs-code-block} yaml
---
data-line-numbers: 2,4-6
---
- name: reboot the servers
  command: /sbin/reboot -t now
- name: restart webserver
  service:
    name: httpd
    state: restarted
```

### Plugins
```{revealjs-code-block} yaml
---
data-line-numbers: 2,6-8
---
- hosts: leaf01
  connection: httpapi
  gather_facts: false
  tasks:
  - name: type a simple arista command
    eos_command:
      commands:
        - show version | json
    register: command_output
---
vars:
  kv: "{{ lookup('consul_kv', 'key', host=localhost) }}"
```

### Collections

```{revealjs-code-block}
---
data-line-numbers: 2-9|10-17
---
namespace/
├── collectionA/
|   ├── docs/
|   ├── galaxy.yml
|   ├── plugins/
|   │   ├── README.md
|   │   └── modules/
|   ├── README.md
|   └── roles/
└── collectionB/
    ├── docs/
    ├── galaxy.yml
    ├── plugins/
    │   ├── connection/
    │   └── modules/
    ├── README.md
    └── roles/
```

```console
$ ansible-galaxy collection
```
