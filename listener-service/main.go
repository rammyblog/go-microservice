package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rammyblog/go-microservice/listener-service/event"
)

func main() {
	// connect to rabbit mq
	rabbitConn, err := connect()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	// listen to msgs
	log.Println("Listening and consuming messages")

	// create a consumer

	consumer, err := event.NewConsumer(rabbitConn)

	if err != nil {
		panic(err)
	}

	// watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})

	if err != nil {
		log.Println(err)
	}

}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second

	var connection *amqp.Connection

	// don't continue until rabbit is ready

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")

		if err != nil {
			fmt.Println("RabbitMQ not ready")
			counts++
		} else {
			connection = c
			log.Println("Conncted to rabbit")

			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("Backing off")
		time.Sleep(backOff)
		continue
	}
	return connection, nil
}
