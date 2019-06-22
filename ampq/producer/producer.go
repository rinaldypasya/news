package producer

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"

	"github.com/rinaldypasya/news/news"
)

var (
	amqpURI = flag.String("amqp", "amqp://guest:guest@localhost:5672/", "AMQP URI")
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func init() {
	flag.Parse()
	initAmqp()
}

var conn *amqp.Connection
var ch *amqp.Channel

func initAmqp() {
	var err error

	conn, err = amqp.Dial(*amqpURI)
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		"test-exchange", // name
		"direct",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // noWait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare the Exchange")
}

// PublishMessage publish messages
func PublishMessages(messages int, news news.News) {
	for i := 0; i < messages; i++ {
		payload, err := json.Marshal(news)
		failOnError(err, "Failed to marshal JSON")

		err = ch.Publish(
			"go-test-exchange", // exchange
			"go-test-key",      // routing key
			false,              // mandatory
			false,              // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Transient,
				ContentType:  "application/json",
				Body:         payload,
				Timestamp:    time.Now(),
			})

		failOnError(err, "Failed to Publish on RabbitMQ")
	}
}
