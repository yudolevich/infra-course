#!/usr/bin/env python3
import sys, pika

if len(sys.argv) < 2:
    print('need message')
    sys.exit()

msg = ' '.join(sys.argv[1:])

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.exchange_declare(exchange='fanout', exchange_type='fanout')
# channel.queue_declare(queue='test')

channel.basic_publish(exchange='fanout', routing_key='test', body=msg)
print(" [x] Sent '{}'".format(msg))
connection.close()
