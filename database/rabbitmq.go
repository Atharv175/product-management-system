package database

import (
	"log"

	"github.com/streadway/amqp"
)

var RabbitConn *amqp.Connection

// ConnectRabbitMQ initializes the connection to RabbitMQ
func ConnectRabbitMQ() {
	var err error
	RabbitConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	log.Println("RabbitMQ connection established")
}

// GetRabbitMQChannel returns a new RabbitMQ channel
func GetRabbitMQChannel() (*amqp.Channel, error) {
	return RabbitConn.Channel()
}

// DeclareImageQueue ensures the image queue exists
func DeclareImageQueue() {
	channel, err := GetRabbitMQChannel()
	if err != nil {
		log.Fatal("Failed to create RabbitMQ channel:", err)
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(
		"image_queue", // Queue name
		true,          // Durable
		false,         // Delete when unused
		false,         // Exclusive
		false,         // No-wait
		nil,           // Arguments
	)
	if err != nil {
		log.Fatal("Failed to declare RabbitMQ queue:", err)
	}
	log.Println("RabbitMQ queue declared")
}
