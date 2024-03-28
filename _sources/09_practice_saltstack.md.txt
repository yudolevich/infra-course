# Salt
В данном практическом занятии познакомимся с базовым использованием
[saltstack][].

## Vagrant
Для работы с [salt][saltstack] будем использовать следующий `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "master" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "master"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      curl -fsSL -o /etc/apt/keyrings/salt-archive-keyring-2023.gpg \
        https://repo.saltproject.io/salt/py3/ubuntu/22.04/amd64/SALT-PROJECT-GPG-PUBKEY-2023.gpg
      echo "deb [signed-by=/etc/apt/keyrings/salt-archive-keyring-2023.gpg arch=amd64] https://repo.saltproject.io/salt/py3/ubuntu/22.04/amd64/latest jammy main" \
        | tee /etc/apt/sources.list.d/salt.list
      apt-get update -q
      apt-get install -yq libnss-mdns salt-master salt-minion
      systemctl enable --now salt-master.service
    SHELL
  end

  config.vm.define "minion1" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "minion1"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      curl -fsSL -o /etc/apt/keyrings/salt-archive-keyring-2023.gpg \
        https://repo.saltproject.io/salt/py3/ubuntu/22.04/amd64/SALT-PROJECT-GPG-PUBKEY-2023.gpg
      echo "deb [signed-by=/etc/apt/keyrings/salt-archive-keyring-2023.gpg arch=amd64] https://repo.saltproject.io/salt/py3/ubuntu/22.04/amd64/latest jammy main" \
        | tee /etc/apt/sources.list.d/salt.list
      apt-get update -q
      apt-get install -yq libnss-mdns salt-minion
      echo 'master: master.local' > /etc/salt/minion.d/master.conf
      systemctl restart salt-minion.service
      systemctl enable salt-minion.service
    SHELL
  end

  config.vm.define "minion2" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "minion2"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      curl -fsSL -o /etc/apt/keyrings/salt-archive-keyring-2023.gpg \
        https://repo.saltproject.io/salt/py3/ubuntu/22.04/amd64/SALT-PROJECT-GPG-PUBKEY-2023.gpg
      echo "deb [signed-by=/etc/apt/keyrings/salt-archive-keyring-2023.gpg arch=amd64] https://repo.saltproject.io/salt/py3/ubuntu/22.04/amd64/latest jammy main" \
        | tee /etc/apt/sources.list.d/salt.list
      apt-get update -q
      apt-get install -yq libnss-mdns salt-minion
      echo 'master: master.local' > /etc/salt/minion.d/master.conf
      systemctl restart salt-minion.service
      systemctl enable salt-minion.service
    SHELL
  end
end
```
Команды будут выполняться на машине `master` из под пользователя `root`.
Для этого после команды `vagrant ssh` можно выполнить команду `sudo -i`.

## Basic Usage

### Masterless
После установки пакетов на машине `master` появится ряд утилит для работы
с [saltstack][]. Для выполнения функций локально на миньоне можно воспользоваться
командой `salt-call`. Чтобы при этом не производилось обращений к мастеру
можно воспользоваться опцией `--local`. Таким образом можно вызвать, например,
функцию `network.ping` без использования мастера:
```console
# salt-call --local network.ping 1.1.1.1
local:
    PING 1.1.1.1 (1.1.1.1) 56(84) bytes of data.
    64 bytes from 1.1.1.1: icmp_seq=1 ttl=63 time=25.9 ms
    64 bytes from 1.1.1.1: icmp_seq=2 ttl=63 time=23.8 ms
    64 bytes from 1.1.1.1: icmp_seq=3 ttl=63 time=23.6 ms
    64 bytes from 1.1.1.1: icmp_seq=4 ttl=63 time=24.0 ms

    --- 1.1.1.1 ping statistics ---
    4 packets transmitted, 4 received, 0% packet loss, time 3005ms
    rtt min/avg/max/mdev = 23.608/24.338/25.946/0.940 ms
```
Переданная функция здесь состоит из имени модуля `network` и имени функции `ping`.

### Documentation
Документацию по функции можно получить передав ее имя функции `sys.doc`:
```console
# salt-call --local sys.doc network.ping
local:
    ----------
    network.ping:

            Performs an ICMP ping to a host

            Changed in version 2015.8.0
                Added support for SunOS

            CLI Example:

                salt '*' network.ping archlinux.org

            New in version 2015.5.0

            Return a True or False instead of ping output.

                salt '*' network.ping archlinux.org return_boolean=True

            Set the time to wait for a response in seconds.

                salt '*' network.ping archlinux.org timeout=3
```
Также можно получить и информацию по всему модулю:
```console
# salt-call --local sys.doc network | head -100
local:
    ----------
    network.active_tcp:

            Return a dict containing information on all of the running TCP connections (currently linux and solaris only)

            Changed in version 2015.8.4

                Added support for SunOS

            CLI Example:

                salt '*' network.active_tcp

    network.arp:

            Return the arp table from the minion

            Changed in version 2015.8.0
                Added support for SunOS

            CLI Example:

                salt '*' network.arp

    network.calc_net:
```

А список доступных модулей и функций можно получить вызовом `sys.list_modules`
и `sys.list_functions` соответственно:
```console
# salt-call --local sys.list_modules | head
local:
    - aliases
    - alternatives
    - archive
    - artifactory
    - baredoc
    - bcache
    - beacons
    - bigip
    - btrfs
# salt-call --local sys.list_functions | head
local:
    - aliases.get_target
    - aliases.has_target
    - aliases.list_aliases
    - aliases.rm_alias
    - aliases.set_target
    - alternatives.auto
    - alternatives.check_exists
    - alternatives.check_installed
    - alternatives.display
```

А также можно просто запустить функцию `sys.doc` без аргументов, чтобы получить
документацию по всем функциям выполняемых модулей.

### Grains
Salt позволяет получить информацию о системе миньона с помощью механизма называемого
[grains][]. Список параметров, которые возможно получить можно посмотреть вызовом
функции `grains.ls`:
```console
# salt-call --local grains.ls | head
local:
    - biosreleasedate
    - biosvendor
    - biosversion
    - boardname
    - cpu_flags
    - cpu_model
    - cpuarch
    - cwd
    - disks
```
Информацию же можно посмотреть функцией `grains.items` для всех параметров или
`grains.item` с указанием конкретного:
```console
# salt-call --local grains.item id
local:
    ----------
    id:
        master
```

## Minions
### Keys
После запуска машин `minion1` и `minion2` они должны были подключиться к мастеру,
но на этом этапе еще нельзя вызывать команды на них, так как не были подтверждены
их ключи. Для просмотра списка ключей можно воспользоваться командой `salt-key -L`:
```console
# salt-key -L
Accepted Keys:
Denied Keys:
Unaccepted Keys:
minion1
minion2
Rejected Keys:
```
Как видно ключи миньонов находятся в `Unaccepted Keys`. Для того чтобы принять
ключи миньонов можно воспользоваться командами `salt-key -a` для конкретного ключа
или `salt-key -A`, чтобы принять все ключи в разделе `Unaccepted Keys`:
```console
# salt-key -a minion1
The following keys are going to be accepted:
Unaccepted Keys:
minion1
Proceed? [n/Y]
Key for minion minion1 accepted.
# salt-key -A
The following keys are going to be accepted:
Unaccepted Keys:
minion2
Proceed? [n/Y]
Key for minion minion2 accepted.
```

### Execution
Теперь выполнять функции на миньонах можно с помощью команды `salt`:
```console
# salt '*' grains.item id
minion2:
    ----------
    id:
        minion2
minion1:
    ----------
    id:
        minion1
```
Здесь мы получили `grains` с информацией о `minion_id`, вторым аргументом команды `salt`
указывается на каких миньонах должна выполниться команда. `*` - означает выполнение на
всех миньонах. Задавать цели для исполнения функции можно различными способами,
например явно указывать имя миньона или с опцией `-G` можно указать его `grains`:
```console
# salt minion1 grains.item ip4_interfaces:enp0s8
minion1:
    ----------
    ip4_interfaces:enp0s8:
        - 192.168.56.42
# salt -G 'ip4_interfaces:enp0s8:0:192.168.56.42' grains.item id
minion1:
    ----------
    id:
        minion1
```

### State
Salt также позволяет описывать состояния в `sls`(Salt State) файлах в `yaml` формате.
По-умолчанию используется каталог `/srv/salt`, в котором находится верхнеуровневый файл
состояния - `top.sls`, который описывает каким миньонам в каких состояниях необходимо
находиться. Создадим директорию `/srv/salt` и файл `top.sls` со следующим содержимым:
```yaml
base:
  '*':
  - nginx
```
Что означает, что ко всем миньоном будет применено состояние описанное в файле `nginx.sls`
в этой же директории. Опишем этот файл:
```yaml
nginx_pkg:        # идентификатор состояния
  pkg.installed:  # функция состояни
  - name: nginx   # аргументы функции
```
Данный файл описывает состояние с установленным пакетом `nginx` на миньонах, для
применения состояния можно воспользоваться командой `salt` с функцией `state.apply`.
Для запуска без внесения изменений для проверки производимых операций можно добавить
аргумент `test=True`:
```console
# salt '*' state.apply test=True
minion2:
----------
          ID: nginx_pkg
    Function: pkg.installed
        Name: nginx
      Result: None
     Comment: The following packages would be installed/updated: nginx
     Started: 21:52:28.189534
    Duration: 101.366 ms
     Changes:
              ----------
              nginx:
                  ----------
                  new:
                      installed
                  old:

Summary for minion2
------------
Succeeded: 1 (unchanged=1, changed=1)
Failed:    0
------------
Total states run:     1
Total run time: 101.366 ms
minion1:
----------
          ID: nginx_pkg
    Function: pkg.installed
        Name: nginx
      Result: None
     Comment: The following packages would be installed/updated: nginx
     Started: 21:52:28.275752
    Duration: 95.124 ms
     Changes:
              ----------
              nginx:
                  ----------
                  new:
                      installed
                  old:

Summary for minion1
------------
Succeeded: 1 (unchanged=1, changed=1)
Failed:    0
------------
Total states run:     1
Total run time:  95.124 ms
```
Данная команда выводит информацию о том какие изменения будут применены. Попробуем
применить состояние:
```console
# salt '*' state.apply
minion1:
----------
          ID: nginx_pkg
    Function: pkg.installed
        Name: nginx
      Result: True
     Comment: The following packages were installed/updated: nginx
     Started: 21:57:15.678622
    Duration: 12706.167 ms
     Changes:
              ----------
              fontconfig-config:
                  ----------
                  new:
                      2.14.1-3ubuntu3
                  old:
              fonts-dejavu-core:
                  ----------
                  new:
                      2.37-6
                  old:
...
Summary for minion2
------------
Succeeded: 1 (changed=1)
Failed:    0
------------
Total states run:     1
Total run time:  13.452 s
```

Добавим собственную конфигурацию `nginx` в `/srv/salt/default.conf`:
```nginx
server {
    listen       80;
    server_name  localhost;

    location / {
        root   /var/www/html;
        index  index.html;
    }
}
```
А также добавим состояние для применения данной конфигурации и перезапуска сервиса
при ее изменении в файл `nginx.sls`:
```yaml
nginx_pkg:        # идентификатор состояния
  pkg.installed:  # функция состояни
  - name: nginx   # аргументы функции

nginx_service:
  service.running:
  - name: nginx
  - reload: True
  - watch:
    - file: /etc/nginx/sites-available/default
  file.managed:
  - name: /etc/nginx/sites-available/default
  - source: salt://default.conf
```
И применим новое состояние:
```console
# salt '*' state.apply
minion2:
----------
          ID: nginx_pkg
    Function: pkg.installed
        Name: nginx
      Result: True
     Comment: All specified packages are already installed
     Started: 22:16:23.382729
    Duration: 25.124 ms
     Changes:
...
minion1:
----------
          ID: nginx_pkg
    Function: pkg.installed
        Name: nginx
      Result: True
     Comment: All specified packages are already installed
     Started: 22:16:23.467122
    Duration: 25.565 ms
     Changes:
----------
          ID: nginx_service
    Function: file.managed
        Name: /etc/nginx/sites-available/default
      Result: True
     Comment: File /etc/nginx/sites-available/default updated
     Started: 22:16:23.495544
    Duration: 42.962 ms
     Changes:
              ----------
              diff:
                  ---
                  +++
                  @@ -1,91 +1,9 @@
                  -##
                  -# You should look at the following URL's in order to grasp a solid understanding
                  -# of Nginx configuration files in order to fully unleash the power of Nginx.
                  -# https://www.nginx.com/resources/wiki/start/
                  -# https://www.nginx.com/resources/wiki/start/topics/tutorials/config_pitfalls/
                  -# https://wiki.debian.org/Nginx/DirectoryStructure
                  -#
                  -# In most cases, administrators will remove this file from sites-enabled/ and
                  -# leave it as reference inside of sites-available where it will continue to be
                  -# updated by the nginx packaging team.
                  -#
                  -# This file will automatically load configuration files provided by other
                  -# applications, such as Drupal or Wordpress. These applications will be made
                  -# available underneath a path with that package name, such as /drupal8.
                  -#
                  -# Please see /usr/share/doc/nginx-doc/examples/ for more detailed examples.
                  -##
                  +server {
                  +    listen       80;
                  +    server_name  localhost;

                  -# Default server configuration
                  -#
                  -server {
                  -     listen 80 default_server;
                  -     listen [::]:80 default_server;
                  -
                  -     # SSL configuration
                  -     #
                  -     # listen 443 ssl default_server;
                  -     # listen [::]:443 ssl default_server;
                  -     #
                  -     # Note: You should disable gzip for SSL traffic.
                  -     # See: https://bugs.debian.org/773332
                  -     #
                  -     # Read up on ssl_ciphers to ensure a secure configuration.
                  -     # See: https://bugs.debian.org/765782
                  -     #
                  -     # Self signed certs generated by the ssl-cert package
                  -     # Don't use them in a production server!
                  -     #
                  -     # include snippets/snakeoil.conf;
                  -
                  -     root /var/www/html;
                  -
                  -     # Add index.php to the list if you are using PHP
                  -     index index.html index.htm index.nginx-debian.html;
                  -
                  -     server_name _;
                  -
                  -     location / {
                  -             # First attempt to serve request as file, then
                  -             # as directory, then fall back to displaying a 404.
                  -             try_files $uri $uri/ =404;
                  -     }
                  -
                  -     # pass PHP scripts to FastCGI server
                  -     #
                  -     #location ~ \.php$ {
                  -     #       include snippets/fastcgi-php.conf;
                  -     #
                  -     #       # With php-fpm (or other unix sockets):
                  -     #       fastcgi_pass unix:/run/php/php7.4-fpm.sock;
                  -     #       # With php-cgi (or other tcp sockets):
                  -     #       fastcgi_pass 127.0.0.1:9000;
                  -     #}
                  -
                  -     # deny access to .htaccess files, if Apache's document root
                  -     # concurs with nginx's one
                  -     #
                  -     #location ~ /\.ht {
                  -     #       deny all;
                  -     #}
                  +    location / {
                  +        root   /var/www/html;
                  +        index  index.html;
                  +    }
                   }
                  -
                  -
                  -# Virtual Host configuration for example.com
                  -#
                  -# You can move that to a different file under sites-available/ and symlink that
                  -# to sites-enabled/ to enable it.
                  -#
                  -#server {
                  -#    listen 80;
                  -#    listen [::]:80;
                  -#
                  -#    server_name example.com;
                  -#
                  -#    root /var/www/example.com;
                  -#    index index.html;
                  -#
                  -#    location / {
                  -#            try_files $uri $uri/ =404;
                  -#    }
                  -#}
----------
          ID: nginx_service
    Function: service.running
        Name: nginx
      Result: True
     Comment: Service reloaded
     Started: 22:16:23.561905
    Duration: 100.347 ms
     Changes:
              ----------
              nginx:
                  True

Summary for minion1
------------
Succeeded: 3 (changed=2)
Failed:    0
------------
Total states run:     3
Total run time: 168.874 ms
```

### Jinja2
Как в передаваемых файлах так и в самих файлах состояний `sls` можно использовать шаблоны,
в качестве движка шаблонизации по-умолчанию используется [`jinja2`][jinja]. Добавим в файл
`nginx.sls` передачу на миньоны файла `index.html` из шаблона:
```yaml
nginx_pkg:        # идентификатор состояния
  pkg.installed:  # функция состояни
  - name: nginx   # аргументы функции

nginx_service:
  service.running:
  - name: nginx
  - reload: True
  - watch:
    - file: /etc/nginx/sites-available/default
  file.managed:
  - name: /etc/nginx/sites-available/default
  - source: salt://default.conf

nginx_index:
  file.managed:
  - name: /var/www/html/index.html
  - source: salt://index.html
  - template: jinja
```
Сам же файл `index.html` добавим в директорию `/srv/salt` со следующим содержимым:
```jinja
hello from {{ grains['id'] }}
```
Как видно в шаблоне мы используем переменную `grains`, содержащая данные, которые мы
наблюдали при вызове функции `grains.items`. Применим новое состояние:
```console
# salt '*' state.apply
minion1:
----------
...
----------
          ID: nginx_index
    Function: file.managed
        Name: /var/www/html/index.html
      Result: True
     Comment: File /var/www/html/index.html updated
     Started: 22:29:12.440232
    Duration: 45.531 ms
     Changes:
              ----------
              diff:
                  New file
              mode:
                  0644

Summary for minion2
------------
Succeeded: 4 (changed=1)
Failed:    0
------------
Total states run:     4
Total run time: 193.739 ms
# curl minion1.local
hello from minion1
# curl minion2.local
hello from minion2
```

### Pillars
Помимо того, что мы можем использовать переменные в `jinja` шаблонах с самих миньонов
через переменную `grains`, мы также можем передавать переменные с мастера на миньоны
через механизм [`pillars`][pillar]. Для этого по умолчанию используется директория
`/srv/pillar`, в которой также должен находиться файл `top.sls`, указывающий каким
миньонам какие переменные необходимо отправлять. Создадим директорию и опишем файл
`top.sls`:
```yaml
base:
  '*':
  - data
```
Таким образом на все миньоны будут отправляться [`pillars`][pillar] из файла `data.sls`
в этой же директории. Зададим содержимое этого файла следующим образом:
```yaml
html:
  - name: 123
    target: minion1
    value: 321
  - name: 2
    target: minion2
    value: hello
```

Теперь попробуем шаблонизировать наш файл состояния `nginx.sls` используя эти переменные:
```jinja
nginx_pkg:        # идентификатор состояния
  pkg.installed:  # функция состояни
  - name: nginx   # аргументы функции

nginx_service:
  service.running:
  - name: nginx
  - reload: True
  - watch:
    - file: /etc/nginx/sites-available/default
  file.managed:
  - name: /etc/nginx/sites-available/default
  - source: salt://default.conf

nginx_index:
  file.managed:
  - name: /var/www/html/index.html
  - source: salt://index.html
  - template: jinja

{% for file in pillar['html'] %}
{% if file['target'] == grains['id'] %}
nginx_html_{{ file['name'] }}:
  file.managed:
  - name: /var/www/html/{{ file['name'] }}
  - contents: {{ file['value'] }}
{% endif %}
{% endfor %}
```
Как видно в шаблоне используется цикл и проверка условия, так что на первом миньоне
должен появиться путь `/123` с содержимым `321`, а на втором путь `/2` с содержимым
`hello`. Применим новое состояние:
```console
# salt '*' state.apply
minion2:
----------
...
----------
          ID: nginx_html_2
    Function: file.managed
        Name: /var/www/html/2
      Result: True
     Comment: File /var/www/html/2 updated
     Started: 22:54:42.659035
    Duration: 2.291 ms
     Changes:
              ----------
              diff:
                  New file

Summary for minion2
------------
Succeeded: 5 (changed=1)
Failed:    0
------------
Total states run:     5
Total run time: 107.908 ms
minion1:
----------
...
----------
          ID: nginx_html_123
    Function: file.managed
        Name: /var/www/html/123
      Result: True
     Comment: File /var/www/html/123 updated
     Started: 22:54:42.682458
    Duration: 2.967 ms
     Changes:
              ----------
              diff:
                  New file

Summary for minion1
------------
Succeeded: 5 (changed=1)
Failed:    0
------------
Total states run:     5
Total run time: 116.716 ms

# curl minion1.local/123
321
# curl minion2.local/2
hello
```




[saltstack]:https://docs.saltproject.io/en/latest/topics/about_salt_project.html#about-salt
[modules]:https://docs.saltproject.io/en/latest/ref/index.html
[grains]:https://docs.saltproject.io/salt/user-guide/en/latest/topics/grains.html
[pillar]:https://docs.saltproject.io/salt/user-guide/en/latest/topics/pillar.html
[jinja]:https://docs.saltproject.io/salt/user-guide/en/latest/topics/jinja.html
