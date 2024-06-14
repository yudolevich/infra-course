# Apache Kafka

В данном практическом занятии познакомимся с базовой работой с брокером сообщений
[Apache Kafka][kafka].

## Vagrant
Для работы с [kafka][] воспользуемся следующим `Vagrantfile`:
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "node" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "node"
    c.vm.network "private_network", type: "dhcp"
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq python3-kafka openjdk-17-jre
      mkdir /opt/kafka
      curl https://dlcdn.apache.org/kafka/3.6.1/kafka_2.13-3.6.1.tgz \
        | tar xz --strip-components=1 -C /opt/kafka
      /opt/kafka/bin/kafka-storage.sh format \
        -t "$(/opt/kafka/bin/kafka-storage.sh random-uuid)" \
        -c /opt/kafka/config/kraft/server.properties
      systemd-run -u kafka -E KAFKA_HEAP_OPTS="-Xmx256M -Xms128M" \
        /opt/kafka/bin/kafka-server-start.sh \
        /opt/kafka/config/kraft/server.properties
      echo 'PATH="$PATH:/opt/kafka/bin"' > /etc/profile.d/kafka.sh
    SHELL
  end
end
```

После развертывания виртуальной машины на ней будет находиться сервер [kafka][],
а также библиотека для взаимодействия из языка `python` - [python-kafka][].

## Status
Вместе с дистрибутивом [apache kafka][kafka] идет набор утилит для работы,
которые находятся в директории `/opt/kafka/bin`. Убедимся в том, что сервер
работает, запустив утилиту `kafka-broker-api-versions.sh`:
```console
$ kafka-broker-api-versions.sh --bootstrap-server localhost:9092
localhost:9092 (id: 1 rack: null) -> (
        Produce(0): 0 to 9 [usable: 9],
        Fetch(1): 0 to 15 [usable: 15],
        ListOffsets(2): 0 to 8 [usable: 8],
        Metadata(3): 0 to 12 [usable: 12],
        LeaderAndIsr(4): UNSUPPORTED,
        StopReplica(5): UNSUPPORTED,
        UpdateMetadata(6): UNSUPPORTED,
        ControlledShutdown(7): UNSUPPORTED,
        OffsetCommit(8): 0 to 8 [usable: 8],
        OffsetFetch(9): 0 to 8 [usable: 8],
        FindCoordinator(10): 0 to 4 [usable: 4],
        JoinGroup(11): 0 to 9 [usable: 9],
        Heartbeat(12): 0 to 4 [usable: 4],
        LeaveGroup(13): 0 to 5 [usable: 5],
        SyncGroup(14): 0 to 5 [usable: 5],
        DescribeGroups(15): 0 to 5 [usable: 5],
        ListGroups(16): 0 to 4 [usable: 4],
        SaslHandshake(17): 0 to 1 [usable: 1],
        ApiVersions(18): 0 to 3 [usable: 3],
        CreateTopics(19): 0 to 7 [usable: 7],
        DeleteTopics(20): 0 to 6 [usable: 6],
        DeleteRecords(21): 0 to 2 [usable: 2],
        InitProducerId(22): 0 to 4 [usable: 4],
        OffsetForLeaderEpoch(23): 0 to 4 [usable: 4],
        AddPartitionsToTxn(24): 0 to 4 [usable: 4],
        AddOffsetsToTxn(25): 0 to 3 [usable: 3],
        EndTxn(26): 0 to 3 [usable: 3],
        WriteTxnMarkers(27): 0 to 1 [usable: 1],
        TxnOffsetCommit(28): 0 to 3 [usable: 3],
        DescribeAcls(29): 0 to 3 [usable: 3],
        CreateAcls(30): 0 to 3 [usable: 3],
        DeleteAcls(31): 0 to 3 [usable: 3],
        DescribeConfigs(32): 0 to 4 [usable: 4],
        AlterConfigs(33): 0 to 2 [usable: 2],
        AlterReplicaLogDirs(34): 0 to 2 [usable: 2],
        DescribeLogDirs(35): 0 to 4 [usable: 4],
        SaslAuthenticate(36): 0 to 2 [usable: 2],
        CreatePartitions(37): 0 to 3 [usable: 3],
        CreateDelegationToken(38): 0 to 3 [usable: 3],
        RenewDelegationToken(39): 0 to 2 [usable: 2],
        ExpireDelegationToken(40): 0 to 2 [usable: 2],
        DescribeDelegationToken(41): 0 to 3 [usable: 3],
        DeleteGroups(42): 0 to 2 [usable: 2],
        ElectLeaders(43): 0 to 2 [usable: 2],
        IncrementalAlterConfigs(44): 0 to 1 [usable: 1],
        AlterPartitionReassignments(45): 0 [usable: 0],
        ListPartitionReassignments(46): 0 [usable: 0],
        OffsetDelete(47): 0 [usable: 0],
        DescribeClientQuotas(48): 0 to 1 [usable: 1],
        AlterClientQuotas(49): 0 to 1 [usable: 1],
        DescribeUserScramCredentials(50): 0 [usable: 0],
        AlterUserScramCredentials(51): 0 [usable: 0],
        DescribeQuorum(55): 0 to 1 [usable: 1],
        AlterPartition(56): UNSUPPORTED,
        UpdateFeatures(57): 0 to 1 [usable: 1],
        Envelope(58): UNSUPPORTED,
        DescribeCluster(60): 0 [usable: 0],
        DescribeProducers(61): 0 [usable: 0],
        UnregisterBroker(64): 0 [usable: 0],
        DescribeTransactions(65): 0 [usable: 0],
        ListTransactions(66): 0 [usable: 0],
        AllocateProducerIds(67): UNSUPPORTED,
        ConsumerGroupHeartbeat(68): UNSUPPORTED
)
```
Команда показывает доступные апи вызовы, которые есть на сервере.

## Send
Для записи сообщений необходимо создать топик, в который мы будем писать. Для
работы с топиком есть утилита `kafka-topics.sh`:
```console
$ kafka-topics.sh --create --topic test \
    --bootstrap-server localhost:9092
Created topic test.
$ kafka-topics.sh --list --bootstrap-server localhost:9092
test
$ kafka-topics.sh --describe --topic test --bootstrap-server localhost:9092
Topic: test     TopicId: Tj0UQ5ZSR4WPDcZNQtbkBA PartitionCount: 1       ReplicationFactor: 1     Configs: segment.bytes=1073741824
        Topic: test     Partition: 0    Leader: 1       Replicas: 1     Isr: 1
```

Для записи в топик можно воспользоваться командой `kafka-console-producer.sh`:
```console
$ kafka-console-producer.sh --topic test --bootstrap-server localhost:9092
>hello
>world
>!
>
```
После записи мы можем посмотреть офсеты в топике командой `kafka-get-offsets.sh`:
```console
$ kafka-get-offsets.sh --bootstrap-server localhost:9092
test:0:3
```

## Receive
Таким образом видно, что в partition 0 топика test есть три сообщения.
Прочитать мы их можем командой `kafka-console-consumer.sh`:
```console
$ kafka-console-consumer.sh --topic test --bootstrap-server localhost:9092 --from-beginning
hello
world
!
^CProcessed a total of 3 messages
```
Сообщения после получения не удаляются и их можно повторно перечитать, также
можно задать `offset` и `partition`:
```console
$ kafka-console-consumer.sh --topic test --bootstrap-server localhost:9092 --partition 0 --offset 1
world
!
^CProcessed a total of 2 messages
```

## Multiple Consumers
Одновременно из топика вычитывать сообщения могут несколько консьюмеров,
для этого можем запустить команду `kafka-console-consumer.sh` в паре
терминалов, а также в отдельном `kafka-console-producer.sh`:
```console
$ kafka-console-consumer.sh --topic test --bootstrap-server localhost:9092
1
2
3
4
5
```
```console
$ kafka-console-consumer.sh --topic test --bootstrap-server localhost:9092
1
2
3
4
5
```
```console
$ kafka-console-producer.sh --topic test --bootstrap-server localhost:9092
>1
>2
>3
>4
>5
>
```
Как видно в таком случае все продюсеры вычитывают одинаковые сообщения.
Для распределения сообщений между консьюмерами необходимо использовать
несколько партиций, создать дополнительные можно командой `kafka-topics.sh`:
```console
$ kafka-topics.sh --alter --topic test --partitions 2 --bootstrap-server localhost:9092
$ kafka-get-offsets.sh --bootstrap-server localhost:9092 | grep test
test:0:8
test:1:0
```

Таким образом мы подключим двух консьюмеров каждого к своей партиции. Для
распределения сообщений можно явно указывать партицию для записи, либо
воспользоваться механизмом распределения по ключам, для этого к каждому
сообщению добавляется ключ и в зависимости от хэша этого ключа будет выбираться
партиция. Запустим консьюмеров в двух терминалах и отдельно продюсера:
```console
$ kafka-console-consumer.sh --topic test --bootstrap-server localhost:9092 --partition 0
hello2
hello5
```
```console
$ kafka-console-consumer.sh --topic test --bootstrap-server localhost:9092 --partition 1
hello1
hello3
hello4
```
```console
$ kafka-console-producer.sh --topic test --bootstrap-server localhost:9092 --property "parse.key=true" --property key.separator=:
>1:hello1
>2:hello2
>3:hello3
>4:hello4
>5:hello5
```

[kafka]:https://kafka.apache.org/
[python-kafka]:https://kafka-python.readthedocs.io/en/master/index.html
