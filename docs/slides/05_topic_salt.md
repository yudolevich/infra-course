## Salt

```{image} ../img/saltstack.svg
:width: 200px
```

### Architecture

```{image} ../img/salt-architecture.png
:width: 450px
```

### Concepts

```{revealjs-fragments}
* Master/Minion
* Modules/States/Formulas
* Grains/Pillar
* Beacons/Reactors
```

### Master/Minion

```{image} ../img/salt-event-system.png
:width: 450px
```

### Modules

```{image} ../img/salt-modules.png
:width: 450px
```

### States

```{image} ../img/salt-states.png
:width: 300px
```

### States

```{image} ../img/salt-states2.png
:width: 450px
```

### Jinja

```jinja
# Declare Jinja list
{% set users = ['fred', 'bob', 'frank'] %}

# Jinja `for` loop
{% for user in users%}
create_{{ user }}:
  user.present:
    - name: {{ user }}
{% endfor %}
```

### Grains

```{image} ../img/salt-grains.png
:width: 300px
```

### Grains

```bash
salt -G 'os:CentOS' test.version
```

```jinja
{{ grains['os'] }}
```

### Pillar

```{image} ../img/salt-pillar.png
:width: 300px
```

### Pillar
```{revealjs-code-block} yaml
---
data-line-numbers: 1-4|5-13
---
/srv/pillar/top.sls:
  base:
    '*':
      - example
/srv/pillar/example.sls:
  pillar1: value
  pillar2:
    - value
    - value
  pillar3:
    sub_key:
      - value
      - value
```

```jinja
{{ pillar['foo']['bar']['baz'] }}
```

### Beacons

```{image} ../img/salt-beacons.png
:width: 450px
```

### Beacons

```{revealjs-code-block} yaml
---
data-line-numbers: 1-10|13-17
---
 beacons:
   inotify:
     - files:
         /etc/named.conf:
            mask:
              - close_write
              - create
              - delete
              - modify
     - disable_during_state_run: True


salt/beacon/20190418-sosf-master/inotify//etc/named.conf {
    "_stamp": "2019-05-06T19:30:35.397508",
    "change": "IN_IGNORED","id": "20190418-sosf-master",
    "path": "/etc/named.conf"
}
```

### Reactors

```{image} ../img/salt-reactors.jpg
:width: 450px
```

### Reactors

```{revealjs-code-block} yaml
---
data-line-numbers: 1-8|10-15
---
/etc/salt/master.d/reactor.conf:
  reactor:
    - 'salt/minion/*/start':
      - /srv/reactor/start.sls
    - 'mycustom/app/tag':
      - /srv/reactor/mycustom.sls
    - salt/beacon/*/inotify//etc/important_file:
      - /srv/reactor/revert.sls
---
/srv/reactor/restart-web-farm.sls:
  restart_service:
    Local.service.restart:
      - tgt: 'web*'
      - arg:
        - httpd
```
