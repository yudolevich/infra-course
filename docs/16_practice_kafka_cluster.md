# Apache Kafka Cluster

В данном практическом занятии рассмотрим работу Consumer Groups, а также
распределение партиций между брокерами в кластере [Apache Kafka][kafka].

## Vagrant
Для работы с [kafka][] воспользуемся следующим `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "broker1" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "broker1"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns python3-kafka openjdk-17-jre
      mkdir /opt/kafka
      curl https://dlcdn.apache.org/kafka/3.6.1/kafka_2.13-3.6.1.tgz \
        | tar xz --strip-components=1 -C /opt/kafka
      sed -i "/^node.id=/s/=.*/=${HOSTNAME: -1}/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^advertised.listeners=/s/localhost/${HOSTNAME}.local/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^controller.quorum.voters=/s/=.*/=1@broker1.local:9093,2@broker2.local:9093,3@broker3.local:9093/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^offsets.topic.replication.factor=/s/=.*/=2/" \
        /opt/kafka/config/kraft/server.properties
      /opt/kafka/bin/kafka-storage.sh format \
        -t "qk89etSXRw6bZhzLg6QWKA" \
        -c /opt/kafka/config/kraft/server.properties
      systemd-run -p Restart=always -u kafka -E KAFKA_HEAP_OPTS="-Xmx256M -Xms128M" \
        /opt/kafka/bin/kafka-server-start.sh \
        /opt/kafka/config/kraft/server.properties
      echo 'PATH="$PATH:/opt/kafka/bin"' > /etc/profile.d/kafka.sh
    SHELL
  end
  config.vm.define "broker2" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "broker2"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns python3-kafka openjdk-17-jre
      mkdir /opt/kafka
      curl https://dlcdn.apache.org/kafka/3.6.1/kafka_2.13-3.6.1.tgz \
        | tar xz --strip-components=1 -C /opt/kafka
      sed -i "/^node.id=/s/=.*/=${HOSTNAME: -1}/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^advertised.listeners=/s/localhost/${HOSTNAME}.local/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^controller.quorum.voters=/s/=.*/=1@broker1.local:9093,2@broker2.local:9093,3@broker3.local:9093/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^offsets.topic.replication.factor=/s/=.*/=2/" \
        /opt/kafka/config/kraft/server.properties
      /opt/kafka/bin/kafka-storage.sh format \
        -t "qk89etSXRw6bZhzLg6QWKA" \
        -c /opt/kafka/config/kraft/server.properties
      systemd-run -p Restart=always -u kafka -E KAFKA_HEAP_OPTS="-Xmx256M -Xms128M" \
        /opt/kafka/bin/kafka-server-start.sh \
        /opt/kafka/config/kraft/server.properties
      echo 'PATH="$PATH:/opt/kafka/bin"' > /etc/profile.d/kafka.sh
    SHELL
  end
  config.vm.define "broker3" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "broker3"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq libnss-mdns python3-kafka openjdk-17-jre
      mkdir /opt/kafka
      curl https://dlcdn.apache.org/kafka/3.6.1/kafka_2.13-3.6.1.tgz \
        | tar xz --strip-components=1 -C /opt/kafka
      sed -i "/^node.id=/s/=.*/=${HOSTNAME: -1}/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^advertised.listeners=/s/localhost/${HOSTNAME}.local/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^controller.quorum.voters=/s/=.*/=1@broker1.local:9093,2@broker2.local:9093,3@broker3.local:9093/" \
        /opt/kafka/config/kraft/server.properties
      sed -i "/^offsets.topic.replication.factor=/s/=.*/=2/" \
        /opt/kafka/config/kraft/server.properties
      /opt/kafka/bin/kafka-storage.sh format \
        -t "qk89etSXRw6bZhzLg6QWKA" \
        -c /opt/kafka/config/kraft/server.properties
      systemd-run -p Restart=always -u kafka -E KAFKA_HEAP_OPTS="-Xmx256M -Xms128M" \
        /opt/kafka/bin/kafka-server-start.sh \
        /opt/kafka/config/kraft/server.properties
      echo 'PATH="$PATH:/opt/kafka/bin"' > /etc/profile.d/kafka.sh
    SHELL
  end
end
```

После развертывания мы получим кластер из трех узлов. Базово операции будем
производить на машине `broker1`.

## Consumer Groups
Для чтения из топика с несколькими партициями множеством консьюмеров существует
удобный механизм Consumer Groups, который позволяет распределить партиции между
консьюмерами в рамках группы.

### Single Consumer
Создадим топик с двумя партициями:
```console
$ kafka-topics.sh --create --topic test --partitions 2 --bootstrap-server localhost:9092
Created topic test.
$ kafka-topics.sh --describe --topic test --bootstrap-server localhost:9092
Topic: test     TopicId: 0AT7xhxMSU6Iu4sXKmGilA PartitionCount: 2       ReplicationFactor: 1    Configs: segment.bytes=1073741824
        Topic: test     Partition: 0    Leader: 3       Replicas: 3     Isr: 3
        Topic: test     Partition: 1    Leader: 1       Replicas: 1     Isr: 1
```
Запустим в отдельном терминале консьюмера, указав для него группу, и отдельно
запустим продюсера, который будет отправлять сообщения по разным партициям:
```console
$ kafka-console-consumer.sh --topic test --group test --bootstrap-server localhost:9092
0
1
2
3
```
```console
$ kafka-console-producer.sh --topic test --property parse.key=true --property key.separator=: --bootstrap-server localhost:9092
>0:0
>1:1
>0:2
>1:3
>
$ kafka-get-offsets.sh --topic test --bootstrap-server localhost:9092
test:0:2
test:1:2
```
Как видно в каждой партиции по 2 сообщения и консьюмер считал их все.

### Multiple Consumers
Попробуем запустить еще одного консьюмера в этой же группе в отдельном терминале,
оставив запущенным старый и запишем еще несколько сообщений:
```console
$ kafka-console-consumer.sh --topic test --group test --bootstrap-server localhost:9092
0
1
2
3
5
7
```
```console
$ kafka-console-consumer.sh --topic test --group test --bootstrap-server localhost:9092
4
6
```
```console
$ kafka-console-producer.sh --topic test --property parse.key=true --property key.separator=: --bootstrap-server localhost:9092
>0:4
>1:5
>0:6
>1:7
>
$ kafka-get-offsets.sh --topic test --bootstrap-server localhost:9092
test:0:4
test:1:4
```
Как видно, после добавления консьюмера в группу сообщения из разных партиций
стали распределяться по разным консьюмерам. Дальнейшее добавление консьюмеров
в группу не даст никакого эффекта, так как в топике всего две партиции, а для
исключения множественной обработки одних данных разными консьюмерами,
предусмотрено чтение из одной партиции только одним консьюмером. Для увеличения
числа консьюмеров необходимо иметь большее число партиций.

### Consumer Offsets
Также для групп консьюмеров создается специальный топик `__consumer_offsets`,
в котором сохраняется информация от консьюмеров об обработанных сообщениях.
Информацию о группе можно получить командой `kafka-consumer-groups.sh`:
```console
$ kafka-consumer-groups.sh --describe --group test --bootstrap-server localhost:9092
GROUP           TOPIC           PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG             CONSUMER-ID                                           HOST            CLIENT-ID
test            test            0          4               4               0               console-consumer-6156fc4c-3e3c-4c25-b818-b43153dc6878 /192.168.56.79  console-consumer
test            test            1          4               4               0               console-consumer-916bf0ae-2b17-44fd-b6ef-1e6b8cd71526 /192.168.56.79  console-consumer
```
Как видно у нас в группе два консьюмера и текущий offset обработанных сообщений
консьюмерами совпадает с концом в партиции. Отключим теперь наших консьюмеров,
если они еще запущены и попробуем записать еще несколько сообщений в топик:
```console
$ kafka-consumer-groups.sh --describe --group test --bootstrap-server localhost:9092
Consumer group 'test' has no active members.

GROUP           TOPIC           PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG             CONSUMER-ID     HOST            CLIENT-ID
test            test            0          4               4               0               -               -               -
test            test            1          4               4               0               -               -               -
$ kafka-console-producer.sh --topic test --property parse.key=true --property key.separator=: --bootstrap-server localhost:9092
>0:0
>1:1
>
$ kafka-conumer-groups.sh --describe --group test: --bootstrap-server localhost:9092
Consumer group 'test' has no active members.

GROUP           TOPIC           PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG             CONSUMER-ID     HOST            CLIENT-ID
test            test            0          4               5               1               -               -               -
test            test            1          4               5               1               -               -               -
```
Теперь в выводе видно, что в партиции топика добавлены еще по одному сообщению
и текущий offset консьюмеров начал отставать. Запустим консьюмер в этой группе
и убедимся, что он вычитает сообщения с заданного offset:
```console
$ kafka-console-consumer.sh --topic test --group test --bootstrap-server localhost:9092
0
1
^CProcessed a total of 2 messages
$ kafka-consumer-groups.sh --describe --group test --bootstrap-server localhost:9092
Consumer group 'test' has no active members.

GROUP           TOPIC           PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG             CONSUMER-ID     HOST            CLIENT-ID
test            test            0          5               5               0               -               -               -
test            test            1          5               5               0               -               -               -
```

## Cluster
[Apache Kafka][kafka] изначально проектировался как распределенная масштабируемая
система, которая работает на множестве узлов и позволяет реплицировать данные
между ними для отказоустойчивости. С помощью `vagrant` мы развернули кластер на
трех узлах, попробуем воспроизвести отказ одного из них и посмотреть на результат.

### Prepare
Информацию о узлах кластера можем посмотреть командой
`kafka-broker-api-versions.sh`:
```console
$ kafka-broker-api-versions.sh --bootstrap-server localhost:9092 | grep '^\w'
broker2.local:9092 (id: 2 rack: null) -> (
broker1.local:9092 (id: 1 rack: null) -> (
broker3.local:9092 (id: 3 rack: null) -> (
```
Для репликации партиций создадим новый топик, указав параметр `replication factor` в 2:
```console
$ kafka-topics.sh --create --topic replicated --partitions 2 --replication-factor 2 --bootstrap-server localhost:9092
Created topic replicated.
$ kafka-topics.sh --describe --topic replicated --bootstrap-server localhost:9092
Topic: replicated       TopicId: O8Z52kMvQWSsOwsJOpoOlw PartitionCount: 2       ReplicationFactor: 2     Configs: segment.bytes=1073741824
        Topic: replicated       Partition: 0    Leader: 1       Replicas: 1,2   Isr: 1,2
        Topic: replicated       Partition: 1    Leader: 2       Replicas: 2,3   Isr: 2,3
```
Как видно, у нас имеется топик `replicated` с двумя партициями, сами же партиции
имеют по две реплики на разных узлах.

### Consume
Запустим пару консьюмеров в разных терминалах, как это делали в предыдущем разделе
и запишем несколько сообщений в топик:
```console
$ kafka-console-consumer.sh --topic replicated --group replicated --bootstrap-server localhost:9092
0
1
```
```console
$ kafka-console-consumer.sh --topic replicated --group replicated --bootstrap-server localhost:9092
2
3
```
```console
$ kafka-console-producer.sh --topic replicated --property parse.key=true --property key.separator=: --bootstrap-server localhost:9092
>0:0
>0:1
>1:2
>1:3
>
```

### Stop Broker
Попробуем отключить второй брокер, так как на нем находится лидер для одной из
партиций, для этого зайдем на машину `broker2` и выполним команду:
```console
$ sudo systemctl stop kafka
```
Вернемся обратно на первый узел и посмотрим состояние:
```console
$ kafka-broker-api-versions.sh --bootstrap-server localhost:9092 | grep '^\w'
broker1.local:9092 (id: 1 rack: null) -> (
broker3.local:9092 (id: 3 rack: null) -> (
$ kafka-topics.sh --describe --topic replicated --bootstrap-server localhost:9092
Topic: replicated       TopicId: O8Z52kMvQWSsOwsJOpoOlw PartitionCount: 2       ReplicationFactor: 2     Configs: segment.bytes=1073741824
        Topic: replicated       Partition: 0    Leader: 1       Replicas: 1,2   Isr: 1
        Topic: replicated       Partition: 1    Leader: 3       Replicas: 2,3   Isr: 3
```
Как видно, узел `broker2` не отображается в списке, а лидером у партиции стал
узел 3.
При этом наши консьюмеры продолжают работать, допишем еще несколько сообщений
в топик, чтобы убедиться:
```console
$ kafka-console-consumer.sh --topic replicated --group replicated --bootstrap-server localhost:9092
0
1
4
6
```
```console
$ kafka-console-consumer.sh --topic replicated --group replicated --bootstrap-server localhost:9092
2
3
5
7
```
```console
$ kafka-console-producer.sh --topic replicated --property parse.key=true --property key.separator=: --bootstrap-server localhost:9092
>0:4
>1:5
>0:6
>1:7
>
```
Таким образом при потери узла работоспособность приложений не нарушилась.


[kafka]:https://kafka.apache.org/
