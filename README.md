# Go game kafka
Is an atempt to make a game client to connect to kafka server and stream the data to all the other clients

## RUN
`go run cmd/client.go`

## RUN kafka

`docker-compose up`

# Cli debug


```
wget http://ftp.unicamp.br/pub/apache/kafka/2.4.0/kafka_2.12-2.4.0.tgz
tar -xvzf kafka_2.12-2.4.0.tgz
```

### CREATE A NEW TOPIC
`./kafka-topics.sh --create --zookeeper localhost:2181 --topic mytopic --partitions 1 --replication-factor 1`

### DESCRIBE TOPIC
`./kafka-topics.sh --describe --zookeeper localhost:2181 --topic mytopic`

### PRODUCER
`./kafka-console-producer.sh --broker-list localhost:29092 --topic mytopic`

### CONSUMER
`./kafka-console-consumer.sh --bootstrap-server localhost:29092 --topic mytopic --from-beginning`