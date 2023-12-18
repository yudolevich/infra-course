## Kafka

```{image} ../img/kafka.svg
:width: 200px
```

### Concepts
```{revealjs-fragments}
* Producer/Consumer
* Cluster/Broker
* Topic/Partitions
* Consumer Groups
```

### Concepts

```{image} ../img/kafka-cluster.svg
:width: 700px
```

### Cluster
![](../img/kafka-cluster1.svg)

```{revealjs-fragments}
* Broker
* Controller(w/o ZooKeeper)
```

### Topic/Partitions
![](../img/kafka-topic.svg)

[ ](../img/kafka-topic.svg)
```{revealjs-fragments}
* Retention
* Replication Factor
* Leader
* Follower
```

### Message

```{revealjs-fragments}
* Value
* Timestamp
* Key
* Headers
```

### Producer
![](../img/kafka-producer.svg)

```{revealjs-fragments}
* acks settings
* partition assignment
```

### Consumer
![](../img/kafka-consumer.svg)
```{revealjs-fragments}
* __consumer_offsets
```
