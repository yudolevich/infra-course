## RabbitMQ

```{image} ../img/rabbitmq.svg
:width: 200px
```

### Concepts

```{revealjs-fragments}
* Publisher/Consumer
* Queue
* Exchange
* Binding
* Messages
```

### AMQP

```{image} ../img/rabbitmq-routing.png
:width: 700px
```

### Publisher

```{image} ../img/rabbitmq-publisher.svg
:width: 700px
```
```{revealjs-fragments}
* Open connection/channel
* Declare entities
* Write exchange
* Define delivery mode/routing key
```

### Exchange/Bindings
```{image} ../img/rabbitmq-exchange.svg
:width: 700px
```

### Exchange Types
```{image} ../img/rabbitmq-exchange.svg
:width: 700px
```
[ ](../img/rabbitmq-exchange.svg)
```{revealjs-fragments}
* Direct
* Fanout
* Topic
* Headers
```

### Exchange Attributes
```{image} ../img/rabbitmq-exchange.svg
:width: 700px
```

```{revealjs-fragments}
* Name
* Durability
* Auto-delete
```

### Queues
```{image} ../img/rabbitmq-queue.svg
:width: 700px
```

Properties
```{revealjs-fragments}
* Name
* Durable
* Exclusive
* Auto-delete
```

### Messages

```{revealjs-fragments}
* Payload
* Routing Key
* Delivery Mode
* Headers
```

### Consumer
```{image} ../img/rabbitmq-consumer.svg
:width: 700px
```

```{revealjs-fragments}
* Open connection/channel
* Declare entities
* Consume from queue
```

### Consume Messages
```{image} ../img/rabbitmq-consumer.svg
:width: 700px
```

```{revealjs-fragments}
* **Push**/Pull API
* Ack/Nack/AutoAck
* Prefetch
```
