package rabbitMq

import (
	"fmt"

	"github.com/streadway/amqp"
)

func consumeOnce() (headers amqp.Table) {
	queueName := "rabbitmq-q"

	connection, err := amqp.Dial("amqp://rabbit:rabbit@localhost:5672")

	if err != nil {
		panic("could not establish connection with RabbitMQ:" + err.Error())
	}

	channel, err := connection.Channel()

	if err != nil {
		panic("could not open RabbitMQ channel:" + err.Error())
	}

	msgs, err := channel.Consume(queueName, "", false, false, false, false, nil)

	if err != nil {
		panic("error consuming the queue: " + err.Error())
	}

	for msg := range msgs {
		fmt.Println("message received: " + string(msg.Body))
		headers = msg.Headers
		msg.Ack(true)
		break
	}

	defer connection.Close()
	return
}
