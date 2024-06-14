# RabbitMQ Routing
В данном практическом занятии различные способы маршрутизации сообщений в
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

## Direct Exchange
![](img/rabbit-practice1.png)

В прошлом практическом занятии использовался exchange типа `fanout`, который
позволял отправлять сообщения сразу всем получателям. Теперь же попробуем
определить для каждого получателя свои сообщения с помощью exchange типа `direct`.
Реализуем на примере отправки логов разной важности - `info`, `warning` и `error`.
Для этого создадим скрипт `send_log.py` для отправки сообщения, который будет
использовать exchange типа `direct`, а в качестве `routing key` - важность
сообщения.

```python
#!/usr/bin/env python
import pika
import sys

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.exchange_declare(exchange='direct_logs', exchange_type='direct')

severity = sys.argv[1] if len(sys.argv) > 1 else 'info'
message = ' '.join(sys.argv[2:]) or 'Hello World!'
channel.basic_publish(
    exchange='direct_logs', routing_key=severity, body=message)
print(f" [x] Sent {severity}:{message}")
connection.close()
```
Данный скрипт позволяет отправлять сообщения, указывая первым аргументом его
важность.

Для получения сообщений создадим еще один скрипт `receive_log.py`, который будет
определять очереди и связывать их с exchange с помощью routing key.
```python
#!/usr/bin/env python
import pika
import sys

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.exchange_declare(exchange='direct_logs', exchange_type='direct')

result = channel.queue_declare(queue='', exclusive=True)
queue_name = result.method.queue

severities = sys.argv[1:]
if not severities:
    sys.stderr.write("Usage: %s [info] [warning] [error]\n" % sys.argv[0])
    sys.exit(1)

for severity in severities:
    channel.queue_bind(
        exchange='direct_logs', queue=queue_name, routing_key=severity)

print(' [*] Waiting for logs. To exit press CTRL+C')


def callback(ch, method, properties, body):
    print(f" [x] {method.routing_key}:{body}")


channel.basic_consume(
    queue=queue_name, on_message_callback=callback, auto_ack=True)

channel.start_consuming()
```
Данный скрипт позволяет создать очередь и с помощью аргументов запуска связать
ее с exchange по нескольким уровням важности сообщений, которые необходимо
получать, а также регистрирует callback, который выводит сообщения на экран.

Запустим в паре терминалов второй скрипт, который будет получать сообщения с
разными уровнями важности, а в отдельном терминале скрипт с отправкой сообщений:
```console
$ python3 receive_log.py info
 [*] Waiting for logs. To exit press CTRL+C
 [x] info:b'test info message'
```
```console
$ python3 receive_log.py info error warning
 [*] Waiting for logs. To exit press CTRL+C
 [x] info:b'test info message'
 [x] warning:b'test warning message'
 [x] error:b'test error message'
```
```console
$ python3 send_log.py info 'test info message'
 [x] Sent info:test info message
$ python3 send_log.py warning 'test warning message'
 [x] Sent warning:test warning message
$ python3 send_log.py error 'test error message'
 [x] Sent error:test error message
```

При запущенных скриптах `receive_log.py` посмотрим информацию с помощью
`rabbitmqctl`:
```console
$ sudo rabbitmqctl list_exchanges | grep direct_logs
direct_logs     direct
$ sudo rabbitmqctl list_queues --quiet
name    messages
amq.gen-oT43COPoX7gTRkNJ2lNdZQ  0
amq.gen-bEcRRYMM1Tl-UOf06Gsk_Q  0
$ sudo rabbitmqctl list_bindings | grep direct_logs
direct_logs     exchange        amq.gen-bEcRRYMM1Tl-UOf06Gsk_Q  queue   error   []
direct_logs     exchange        amq.gen-bEcRRYMM1Tl-UOf06Gsk_Q  queue   info    []
direct_logs     exchange        amq.gen-oT43COPoX7gTRkNJ2lNdZQ  queue   info    []
direct_logs     exchange        amq.gen-bEcRRYMM1Tl-UOf06Gsk_Q  queue   warning []
```
Как видно из вывода в [rabbitmq][] имеется exchange с именем `direct_logs` и типом
`direct`, две очереди созданные запущенными скриптами `receive_log.py` и набор
`bindings`, которые связывают exchange `direct_logs` с очередями с помощью
`routing keys` с именами `info`, `warning` и `error`.

## Topics
![](img/rabbit-practice2.png)

Помимо связывания exchange с очередью с помощью простого routing key, в [rabbitmq
][] имеется более гибкий exchange типа `topic`, который позволяет использовать в
качестве ключа список слов разделенных точками. А также позволяет использовать
специальные символы:
- `*` - может заменить одно слово
- `#` - может заменить любое количество слов
Таким образом в качестве routing key могут использоваться различные варианты:
`*.orange.*`, `*.*.rabbit`, `lazy.#`.

Изменим наш скрипт `send_log.py` для использования exchange типа `topic`:
```python
#!/usr/bin/env python
import pika
import sys

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.exchange_declare(exchange='topic_logs', exchange_type='topic')

routing_key = sys.argv[1] if len(sys.argv) > 2 else 'anonymous.info'
message = ' '.join(sys.argv[2:]) or 'Hello World!'
channel.basic_publish(
    exchange='topic_logs', routing_key=routing_key, body=message)
print(f" [x] Sent {routing_key}:{message}")
connection.close()
```
Логика остается такой же, только теперь мы можем указывать routing key состоящий
из нескольких слов разделенных точкой.

Также изменим скрипт `receive_log.py`:
```python
#!/usr/bin/env python
import pika
import sys

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.exchange_declare(exchange='topic_logs', exchange_type='topic')

result = channel.queue_declare('', exclusive=True)
queue_name = result.method.queue

binding_keys = sys.argv[1:]
if not binding_keys:
    sys.stderr.write("Usage: %s [binding_key]...\n" % sys.argv[0])
    sys.exit(1)

for binding_key in binding_keys:
    channel.queue_bind(
        exchange='topic_logs', queue=queue_name, routing_key=binding_key)

print(' [*] Waiting for logs. To exit press CTRL+C')


def callback(ch, method, properties, body):
    print(f" [x] {method.routing_key}:{body}")


channel.basic_consume(
    queue=queue_name, on_message_callback=callback, auto_ack=True)

channel.start_consuming()
```

Как и в предыдущем разделе запустим в двух терминала `receive_log.py` и в
отдельном `send_log.py`:
```console
$ python3 receive_log.py *.info test.*
 [*] Waiting for logs. To exit press CTRL+C
 [x] test.info:b'test message'
 [x] dev.info:b'test message'
 [x] test.one:b'test message'
```
```console
$ python3 receive_log.py 'test.#'
 [*] Waiting for logs. To exit press CTRL+C
 [x] test.one:b'test message'
 [x] test.one.two:b'test message'
```
```console
$ python3 send_log.py test.info 'test message'
 [x] Sent test.info:test message
$ python3 send_log.py dev.info 'test message'
 [x] Sent dev.info:test message
$ python3 send_log.py test.one 'test message'
 [x] Sent test.one:test message
$ python3 send_log.py test.one.two 'test message'
 [x] Sent test.one.two:test message
```

При запущенных скриптах `receive_log.py` посмотрим информацию с помощью
`rabbitmqctl`:
```console
$ sudo rabbitmqctl list_exchanges | grep topic_logs
topic_logs      topic
$ sudo rabbitmqctl list_queues --quiet
name    messages
amq.gen-HHnB6I713X9fiH8cyM_RCA  0
amq.gen-yryF4eJWFgBMSTpPPEmHOQ  0
$ sudo rabbitmqctl list_bindings | grep topic_logs
topic_logs      exchange        amq.gen-HHnB6I713X9fiH8cyM_RCA  queue   *.info  []
topic_logs      exchange        amq.gen-yryF4eJWFgBMSTpPPEmHOQ  queue   test.#  []
topic_logs      exchange        amq.gen-HHnB6I713X9fiH8cyM_RCA  queue   test.*  []
```
Как видно вывод аналогичен exchange с типом `direct`, но в качестве routing key
используются более гибкие варианты ключей.

## RPC
![](img/rabbit-practice3.png)

С помощью [rabbitmq][] также возможно реализовать rpc(remote procedure call), это
можно реализовать следующей последовательностью:
- Клиент при старте создает анонимную эксклюзивную очередь для ответов.
- Для rpc запросов клиент отправляет сообщение с двумя параметрами:
  `reply_to` с указанием очереди ответов и `correlation_id` с уникальным
  идентификатором запроса.
- Запрос отправляется в очередь `rpc_queue`.
- RPC сервер ожидает запросы в этой очереди, при получении выполняет обработку
  и возвращает обратно клиенту в очередь из параметра `reply_to`.
- Клиент ожидает сообщения в очереди для ответов, валидирует `correlation_id` и
  возвращает ответ в приложение.

Создадим скрипт `rpc_server.py`, который будет получать число из очереди и
вычислять для него значение в ряду Фибоначчи:
```python
#!/usr/bin/env python
import pika

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost'))

channel = connection.channel()

channel.queue_declare(queue='rpc_queue')

def fib(n):
    if n == 0:
        return 0
    elif n == 1:
        return 1
    else:
        return fib(n - 1) + fib(n - 2)

def on_request(ch, method, props, body):
    n = int(body)

    print(f" [.] fib({n})")
    response = fib(n)

    ch.basic_publish(exchange='',
                     routing_key=props.reply_to,
                     properties=pika.BasicProperties(correlation_id = \
                                                         props.correlation_id),
                     body=str(response))
    ch.basic_ack(delivery_tag=method.delivery_tag)

channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue='rpc_queue', on_message_callback=on_request)

print(" [x] Awaiting RPC requests")
channel.start_consuming()
```
А также создадим клиента, который будет вызывать rpc на сервере:
```python
#!/usr/bin/env python
import pika
import uuid
import sys


class FibonacciRpcClient(object):

    def __init__(self):
        self.connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='localhost'))

        self.channel = self.connection.channel()

        result = self.channel.queue_declare(queue='', exclusive=True)
        self.callback_queue = result.method.queue

        self.channel.basic_consume(
            queue=self.callback_queue,
            on_message_callback=self.on_response,
            auto_ack=True)

        self.response = None
        self.corr_id = None

    def on_response(self, ch, method, props, body):
        if self.corr_id == props.correlation_id:
            self.response = body

    def call(self, n):
        self.response = None
        self.corr_id = str(uuid.uuid4())
        self.channel.basic_publish(
            exchange='',
            routing_key='rpc_queue',
            properties=pika.BasicProperties(
                reply_to=self.callback_queue,
                correlation_id=self.corr_id,
            ),
            body=str(n))
        self.connection.process_data_events(time_limit=None)
        return int(self.response)


fibonacci_rpc = FibonacciRpcClient()

n = sys.argv[1] if len(sys.argv) > 1 else 1

print(f" [x] Requesting fib({n})")
response = fibonacci_rpc.call(n)
print(f" [.] Got {response}")
```

Теперь запустим в разных терминалах сервер и клиент:
```console
$ python3 rpc_server.py
 [x] Awaiting RPC requests
 [.] fib(1)
 [.] fib(10)
 [.] fib(30)
```
```console
$ python3 rpc_client.py 1
 [x] Requesting fib(1)
 [.] Got 1
$ python3 rpc_client.py 10
 [x] Requesting fib(10)
 [.] Got 55
$ python3 rpc_client.py 30
 [x] Requesting fib(30)
 [.] Got 832040
```
Как видно, используя [rabbitmq][] нам удалось реализовать RPC.



[rabbitmq]:https://www.rabbitmq.com/
[pika]:https://pika.readthedocs.io/en/stable/
