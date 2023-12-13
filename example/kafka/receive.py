import sys
from kafka import KafkaConsumer, TopicPartition

if len(sys.argv) < 3:
    print(f'usage: {sys.argv[0]} <topic> <partition>')
    sys.exit(1)

consumer = KafkaConsumer(bootstrap_servers='localhost:9092')
consumer.assign([TopicPartition(sys.argv[1], int(sys.argv[2]))])
for msg in consumer:
    print(msg.value)
