nginx:
  pkg.installed: []

pyinotify:
  pip.installed: []

/var/www/html/index.html:
  file.managed:
  - contents: |
      hello from {{ grains['id'] }}

beacon_index:
  beacon.present:
  - save: True
  - enable: True
  - files:
      /var/www/html/index.html:
        mask:
        - modify
  - disable_during_state_run: True
  - beacon_module: inotify
