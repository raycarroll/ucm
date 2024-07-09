package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}

func MoveFile(source, destination string) (err error) {
	return os.Rename(source, destination)
}

func readSpectrumFiles(spectrumPath string) []string {

	file, err := os.Open(spectrumPath)
	var fileList []string

	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	entries, err := file.ReadDir(0)

	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		fileList = append(fileList, e.Name())
	}

	return fileList
}

func archiveDataFile(dataFile string) {
	//move data file from current directory to /processed directory

}

func readFile(path string, filename string) (string, error) {
	body, err := ioutil.ReadFile(path + filename)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
		return "", err
	}
	return string(body), nil
}

func isDir(filepath string) bool {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	return fi.IsDir()
}
func send() {
	//username := "default_user_hU_V7zaOVkqdnAb96no"
	//password := "UK5IoTeB7hD6tf3BGfV4ewt9hs0Dw3WN"
	username := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)

	println("user ", username)
	println("pass ", password)
	println("url ", url)

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filePath := "/docker/starlight/shared_directory/config_files_starlight/spectrum/"

	for {
		dataFiles := readSpectrumFiles(filePath)
		log.Printf("Num Files = %v", len(dataFiles))

		for _, f := range dataFiles {
			//for each datafile, dreate event and move file to /processed directory

			if !isDir(filePath + f) {
				log.Printf("File = %s", f)

				datafile, err := readFile("/docker/starlight/shared_directory/config_files_starlight/spectrum/", f)

				body := datafile
				headers := make(amqp.Table)
				headers["filename"] = f

				err = ch.PublishWithContext(ctx,
					"",     // exchange
					q.Name, // routing key
					false,  // mandatory
					false,  // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(body),
						Headers:     headers,
					})
				failOnError(err, "Failed to publish a message")
				log.Printf(" [x] Sent %s\n", body[0:10])

				//move data file to processed
				log.Printf(" Moving file to /processed dir")

				moveerr := MoveFile(filePath+f, filePath+"processed/"+f)

				if moveerr != nil {
					log.Printf("unable to move file: %v", moveerr)
				}

				log.Printf(" Moved file to " + filePath + "/processed/" + f)
			}
		}

		// wait x seconds to read data directory again
		time.Sleep(60 * time.Second)
	}

}
