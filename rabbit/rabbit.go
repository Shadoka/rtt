package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"rtt/data"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// this verbose functionality is copied from root.go to prevent cycles

var Verbose bool

func VerbosePrintln(text string) {
	if Verbose {
		fmt.Println(text)
	}
}

func Connect(host string, port string, user string, password string) (conn *amqp.Connection) {
	connectionString := fmt.Sprintf("amqp://%v:%v@%v:%v", user, password, host, port)
	VerbosePrintln(fmt.Sprintf("Connecting to rabbitmq with connection string '%v'", connectionString))

	var err error
	if conn, err = amqp.Dial(connectionString); err != nil {
		log.Fatalf("Unable to connect to rabbitmq: %v", err)
	}

	return
}

func GetChannel(connection *amqp.Connection) *amqp.Channel {
	ch, err := connection.Channel()
	if err != nil {
		log.Fatalf("Unable to create channel with rabbitmq: %v", err)
	}
	return ch
}

func DeclareQueue(queueInfo *data.RabbitQueue, ch *amqp.Channel) amqp.Queue {
	queue, err := ch.QueueDeclare(
		queueInfo.Name,
		queueInfo.Durable,
		queueInfo.AutoDelete,
		queueInfo.Exclusive,
		queueInfo.NoWait,
		queueInfo.AmqpTable)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to get or create queue: %v", err)
		os.Exit(1)
	}

	VerbosePrintln(fmt.Sprintf("Successfully created queue with name '%v'", queueInfo.Name))
	return queue
}

func GetPassiveQueue(queueName string, ch *amqp.Channel) amqp.Queue {
	// only queue name matters
	queue, err := ch.QueueDeclarePassive(
		queueName,
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to get queue: %v", err)
		os.Exit(1)
	}

	return queue
}

func BindQueue(queueInfo *data.RabbitQueue, ch *amqp.Channel) {
	err := ch.QueueBind(queueInfo.Name,
		queueInfo.Key,
		queueInfo.Exchange,
		queueInfo.NoWait,
		nil)
	if err != nil {
		log.Fatalf("unable to bind queue '%v' with key '%v' to exchange '%v': %v",
			queueInfo.Name,
			queueInfo.Key,
			queueInfo.Exchange,
			err)
	}
	VerbosePrintln(fmt.Sprintf("Binding created: %v ==(%v)==> %v", queueInfo.Exchange, queueInfo.Key, queueInfo.Name))
}

func DeclareExchange(exchangeInfo data.RabbitExchange, ch *amqp.Channel) {
	err := ch.ExchangeDeclare(exchangeInfo.Name,
		exchangeInfo.Kind,
		exchangeInfo.Durable,
		exchangeInfo.AutoDelete,
		exchangeInfo.Internal,
		exchangeInfo.NoWait,
		exchangeInfo.AmqpTable)
	if err != nil {
		log.Fatalf("unable to create exchange '%v': %v", exchangeInfo.Name, err)
	}
	VerbosePrintln(fmt.Sprintf("Successfully created exchange with name '%v'", exchangeInfo.Name))
}

func SendMessage(data data.InputQueue, channel *amqp.Channel) {
	msgPayload, err := json.Marshal(data.Data)
	if err != nil {
		log.Fatalf("Unable to create message payload: %v", err)
	}

	msgId := uuid.New().String()
	err = channel.Publish(data.Queue.Exchange,
		data.Queue.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        msgPayload,
			MessageId:   msgId,
		})
	if err != nil {
		log.Fatalf("Unable to publish message in queue '%v': %v", data.Queue.Name, err)
	}
}

func CreateConsumer(channel *amqp.Channel, queueInfo *data.RabbitQueue) <-chan amqp.Delivery {
	msgs, err := channel.Consume(queueInfo.Name,
		uuid.NewString(),
		true,
		queueInfo.Exclusive,
		false, // noLocal unsupported by RabbitMQ
		queueInfo.NoWait,
		nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot create consumer: %v", err)
		os.Exit(1)
	}
	return msgs
}

func CreateConsumerWithConfiguration(channel *amqp.Channel, queueInfo *data.RabbitQueue, ack bool, ctx context.Context) <-chan amqp.Delivery {
	msgs, err := channel.ConsumeWithContext(ctx,
		queueInfo.Name,
		uuid.NewString(),
		ack,
		queueInfo.Exclusive,
		false, // noLocal unsupported by RabbitMQ
		queueInfo.NoWait,
		nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot create consumer: %v", err)
		os.Exit(1)
	}
	return msgs
}
