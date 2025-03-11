package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	log.Printf(" ------------------ receive() --------------------- ")
	username := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)

	log.Printf("Connecting to RabbitMQ with URL: %s", url)

	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"starlight", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		log.Printf(" -------------------------- Iterating messages  -------------------------------")
		for d := range msgs {
			log.Printf(" -------------------------- Received a message: %s -------------------------------", d.Headers["filename"])

			if filename, ok := d.Headers["filename"].(string); ok {

				if strings.HasSuffix(filename, ".in") {
					// If it's an .in file, add it to the process list and acknowledge the message
					updateToProcessList(filename)
					d.Ack(false)
					continue
				}

				// Write the received file to the output directory
				outputPath := os.Getenv("OUTPUT_DIR")
				if outputPath == "" {
					log.Fatal("OUTPUT_DIR environment variable is required")
				}

				// Ensure the output directory exists
				if exists, _ := exists(outputPath); !exists {
					log.Printf("Output directory does not exist, creating it: %s", outputPath)
					err := os.Mkdir(outputPath, 0700)
					if err != nil {
						log.Printf("Error creating output directory: %v", err)
					}
				}

				// Write the file content to the output directory
				filePath := filepath.Join(outputPath, filename)
				err := os.WriteFile(filePath, d.Body, 0644)
				if err != nil {
					log.Printf("Error writing file %s: %v", filename, err)
				} else {
					log.Printf("Successfully wrote file %s to %s", filename, outputPath)
				}

				// Acknowledge the message
				d.Ack(false)
			} else {
				log.Printf("Error: filename header is missing or invalid")
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

// updateToProcessList adds the .in file to the processing list
func updateToProcessList(inFileName string) {
	log.Printf("Adding new .in file to ProcessList %s", inFileName)
	PROCESS_LIST := os.Getenv("PROCESS_LIST")

	if err := touchFile(PROCESS_LIST); err != nil {
		log.Printf("Error creating process list file: %v", err)
	}

	// Append the .in file to the process list
	f, err := os.OpenFile(PROCESS_LIST, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Printf("Error opening process list file: %v", err)
		return
	}
	defer f.Close()

	//add new .in file to bottom of list

	if _, err = f.WriteString(inFileName + "\n"); err != nil {
		log.Printf("Error writing to process list file: %v", err)
	} else {
		log.Printf("Added %s to process list", inFileName)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
