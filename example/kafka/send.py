import sys
from kafka import KafkaProducer, KafkaAdminClient
from kafka.admin.new_partitions import NewPartitions

if len(sys.argv) < 4:
    print(f'usage: {sys.argv[0]} <topic> <partition> <message>')
    sys.exit(1)

producer = KafkaProducer(bootstrap_servers='localhost:9092')
if int(sys.argv[2]) not in producer.partitions_for(sys.argv[1]):
    client = KafkaAdminClient(bootstrap_servers='localhost:9092')
    rslt = client.create_partitions({
        sys.argv[1]: NewPartitions(int(sys.argv[2])+1),
    })
    producer = KafkaProducer(bootstrap_servers='localhost:9092')

producer.send(sys.argv[1], partition=int(sys.argv[2]),
              value=bytes(sys.argv[3], encoding='utf-8')
              ).get(timeout=10)
