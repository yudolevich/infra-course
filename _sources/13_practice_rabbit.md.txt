# RabbitMQ
В данном практическом занятии познакомимся с базовой работой с брокером сообщений
[RabbitMQ][].

## Vagrant
Для работы с [rabbitmq][] воспользуемся следующим `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "node" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "node"
    c.vm.network "forwarded_port", guest: 15672, host: 15672
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq python3-pip rabbitmq-server
      rabbitmq-plugins enable rabbitmq_management
      rabbitmqctl add_user admin admin
      rabbitmqctl set_user_tags admin administrator
      rabbitmqctl set_permissions -p / admin ".*" ".*" ".*"
      pip3 install pika --break-system-packages
    SHELL
  end
end
```

После развертывания виртуальной машины на ней будет находиться сервер [rabbitmq][],
в нем создастся пользователь `admin` и наружу будет выставлен
[веб интерфейс](http://localhost:15672), а также установится библиотека для
взаимодействия из языка `python` - [pika][].

## Status
Взаимодействовать с сервером [rabbitmq][] можно утилитой `rabbitmqctl`, либо
посредством [веб интерфейса](http://localhost:15672). Информацию о сервере
можно получить с помощью подкоманды `status`:
```console
$ sudo rabbitmqctl status | tail -15

Low free disk space watermark: 0.05 gb
Free disk space: 39.1751 gb

Totals

Connection count: 0
Queue count: 0
Virtual host count: 1

Listeners

Interface: [::], port: 25672, protocol: clustering, purpose: inter-node and CLI tool communication
Interface: [::], port: 5672, protocol: amqp, purpose: AMQP 0-9-1 and AMQP 1.0
Interface: [::], port: 15672, protocol: http, purpose: HTTP API
```
Как видно из вывода на текущий момент нет активных соединений и очереди
отсутствуют. Список очередей можно посмотреть подкомандой `list_queues`:
```console
$ sudo rabbitmqctl list_queues
Timeout: 60.0 seconds ...
Listing queues for vhost / ...
```

## Send
Создадим скрипт на `python`, который создаст очередь при ее отсутствии и добавит
в нее сообщение, переданное через аргументы запуска:
```python
#!/usr/bin/env python3
import sys, pika

if len(sys.argv) < 2:
    print('need message')
    sys.exit()

msg = ' '.join(sys.argv[1:])

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.queue_declare(queue='test')

channel.basic_publish(exchange='', routing_key='test', body=msg)
print(" [x] Sent '{}'".format(msg))
connection.close()
```
Как видно в данном скрипте мы декларируем очередь с именем `test` и публикуем
в нее сообщение. Сохраним данный скрипт в файле `send.py` и запустим его:
```console
$ python3 send.py test
 [x] Sent 'test'
```

Если теперь посмотреть список очередей, то можно увидеть:
```console
$ sudo rabbitmqctl list_queues
Timeout: 60.0 seconds ...
Listing queues for vhost / ...
name    messages
test    1
```
Появилась очередь `test` с одним сообщением в ней.

## Receive
Создадим также скрипт, который будет получать сообщение из этой же очереди:
```python
#!/usr/bin/env python3
import pika, sys

def main():
    connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
    channel = connection.channel()

    channel.queue_declare(queue='test')

    def callback(ch, method, properties, body):
        print(f" [x] Received {body}")

    channel.basic_consume(queue='test', on_message_callback=callback, auto_ack=True)

    print(' [*] Waiting for messages. To exit press CTRL+C')
    channel.start_consuming()

if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        pass
```
Сохраним его в файле `receive.py` и запустим:
```console
$ python3 receive.py
 [*] Waiting for messages. To exit press CTRL+C
 [x] Received b'test'
^C
$ sudo rabbitmqctl list_queues
Timeout: 60.0 seconds ...
Listing queues for vhost / ...
name    messages
test    0
```
Мы получили сообщение из очереди, и, как видно из вывода команды `list_queues`,
сообщение было из очереди удалено, так как в скрипте у нас выставлен параметр
`auto_ack=True`(автоматическое подтверждение получения).

## FIFO
Обработка очереди происходит в том же порядке, что и добавление по принципу FIFO.
Запустим скрипт `send.py` в цикле для добавления в очередь нескольких сообщений:
```console
$ for i in {1..10};do python3 send.py test $i;done
 [x] Sent 'test 1'
 [x] Sent 'test 2'
 [x] Sent 'test 3'
 [x] Sent 'test 4'
 [x] Sent 'test 5'
 [x] Sent 'test 6'
 [x] Sent 'test 7'
 [x] Sent 'test 8'
 [x] Sent 'test 9'
 [x] Sent 'test 10'
$ python3 receive.py
 [*] Waiting for messages. To exit press CTRL+C
 [x] Received b'test 1'
 [x] Received b'test 2'
 [x] Received b'test 3'
 [x] Received b'test 4'
 [x] Received b'test 5'
 [x] Received b'test 6'
 [x] Received b'test 7'
 [x] Received b'test 8'
 [x] Received b'test 9'
 [x] Received b'test 10'
^C
```
Как видно обработка произошла в той же последовательности, что и добавление.

## Multiple Consumers
В одной очереди может быть несколько получателей между которыми будут
распределяться сообщения, запустим скрипт `receive.py` в паре разных терминалов,
а в отдельном цикл с `send.py`:

```console
$ for i in {1..10};do python3 send.py test $i;done
 [x] Sent 'test 1'
 [x] Sent 'test 2'
 [x] Sent 'test 3'
 [x] Sent 'test 4'
 [x] Sent 'test 5'
 [x] Sent 'test 6'
 [x] Sent 'test 7'
 [x] Sent 'test 8'
 [x] Sent 'test 9'
 [x] Sent 'test 10'
```

```console
$ python3 receive.py
 [*] Waiting for messages. To exit press CTRL+C
 [x] Received b'test 1'
 [x] Received b'test 3'
 [x] Received b'test 5'
 [x] Received b'test 7'
 [x] Received b'test 9'
```

```console
$ python3 receive.py
 [*] Waiting for messages. To exit press CTRL+C
 [x] Received b'test 2'
 [x] Received b'test 4'
 [x] Received b'test 6'
 [x] Received b'test 8'
 [x] Received b'test 10'
```

Для того, чтобы одновременно отправить сообщение нескольким потребителям
необходимо изменить тип `exchange` на `fanout` и сделать отдельные очереди для
каждого потребителя и `binding` к ним. Добавим в наши скрипты
`send.py` и `receive.py`:
```python
#!/usr/bin/env python3
import sys, pika

if len(sys.argv) < 2:
    print('need message')
    sys.exit()

msg = ' '.join(sys.argv[1:])

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

# Декларация exchange с типом fanout
channel.exchange_declare(exchange='fanout', exchange_type='fanout')
# channel.queue_declare(queue='test')

channel.basic_publish(exchange='fanout', routing_key='test', body=msg)
print(" [x] Sent '{}'".format(msg))
connection.close()
```

```python
#!/usr/bin/env python3
import pika, sys

def main(queue):
    connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
    channel = connection.channel()

    # Декларация exchange с типом fanout
    channel.exchange_declare(exchange='fanout', exchange_type='fanout')
    channel.queue_declare(queue=queue)

    def callback(ch, method, properties, body):
        print(f" [x] Received {body}")

    # Создание связи exchange с конкретной очередью
    channel.queue_bind(exchange='fanout', queue=queue)
    channel.basic_consume(queue=queue, on_message_callback=callback, auto_ack=True)

    print(' [*] Waiting for messages. To exit press CTRL+C')
    channel.start_consuming()

if __name__ == '__main__':
    try:
        if len(sys.argv) < 2:
            sys.exit()
        main(sys.argv[1])
    except KeyboardInterrupt:
        pass
```

И также запустим в разных терминалах:
```console
$ for i in {1..10};do python3 send.py test $i;done
 [x] Sent 'test 1'
 [x] Sent 'test 2'
 [x] Sent 'test 3'
 [x] Sent 'test 4'
 [x] Sent 'test 5'
 [x] Sent 'test 6'
 [x] Sent 'test 7'
 [x] Sent 'test 8'
 [x] Sent 'test 9'
 [x] Sent 'test 10'
```

```console
$ python3 receive.py test1
 [*] Waiting for messages. To exit press CTRL+C
 [x] Received b'test 1'
 [x] Received b'test 2'
 [x] Received b'test 3'
 [x] Received b'test 4'
 [x] Received b'test 5'
 [x] Received b'test 6'
 [x] Received b'test 7'
 [x] Received b'test 8'
 [x] Received b'test 9'
 [x] Received b'test 10'
```

```console
$ python3 receive.py test2
 [*] Waiting for messages. To exit press CTRL+C
 [x] Received b'test 1'
 [x] Received b'test 2'
 [x] Received b'test 3'
 [x] Received b'test 4'
 [x] Received b'test 5'
 [x] Received b'test 6'
 [x] Received b'test 7'
 [x] Received b'test 8'
 [x] Received b'test 9'
 [x] Received b'test 10'
```


[rabbitmq]:https://www.rabbitmq.com/
[pika]:https://pika.readthedocs.io/en/stable/
