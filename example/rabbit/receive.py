#!/usr/bin/env python3
import pika, sys

def main(queue):
    connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
    channel = connection.channel()

    channel.exchange_declare(exchange='fanout', exchange_type='fanout')
    channel.queue_declare(queue=queue)

    def callback(ch, method, properties, body):
        print(f" [x] Received {body}")

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
