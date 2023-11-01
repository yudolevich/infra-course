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
