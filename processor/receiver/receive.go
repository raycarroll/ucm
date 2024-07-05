package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	log.Printf(" ------------------ reveive() --------------------- ")
	//username := "default_user_hU_V7zaOVkqdnAb96no"
	//password := "UK5IoTeB7hD6tf3BGfV4ewt9hs0Dw3WN"
	username := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	//host := os.Getenv("RABBITMQ_HOST")
	//port := os.Getenv("RABBITMQ_PORT")
	url := "amqp://" + "guest" + ":" + "guest" + "@" + "172.30.198.237" + ":" + "5672" + "/"

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
		log.Println("Read Flag = ", checkReadFlag())

		for {

			if checkReadFlag() == false {

				log.Println("Read Flag = ", checkReadFlag())
				d := <-msgs

				//for d := range msgs {
				log.Printf(" -------------------------- Received a message: %s -------------------------------", d.Headers["filename"])

				//d.Ack(false)

				//filename := string(d.Headers["filename"])

				if filename, ok := d.Headers["filename"].(string); ok {
					writeMessageToFile(d.Body, filename)
					updateInFile(filename)
					setReadFlag()
					d.Ack(false)
				} else {
					log.Println("Error ", ok)
				}
			}
		}
		log.Printf("Processed Messages")
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func writeMessageToFile(body []byte, filename string) {
	println("Writing message to data file")
	filepath := "/starlight/input/"

	if exists, direrr := exists(filepath); exists == false && direrr == nil {
		println("Input directory does not exist, making it", filepath)
		err := os.Mkdir(filepath, 0700)
		if err != nil {
			println("ERROR: ", err)
		}
	}
	err := os.WriteFile(filepath+filename, body, 0644)
	if err != nil {
		println("Error ", err.Error())
	}
}

func checkReadFlag() bool {
	// checks for start_starlight file
	DATA_FILE_FLAG := "/starlight/start_starlight"

	if _, err := os.Stat(DATA_FILE_FLAG); err == nil {
		//log.Println("File "+fi.Name())
		return true
	}
	return false
}
func setReadFlag() {
	// creates the start_starlight file
	log.Printf("Setting the read flag")
	DATA_FILE_FLAG := "/starlight/start_starlight"

	d1 := []byte("start")
	err := os.WriteFile(DATA_FILE_FLAG, d1, 0666)
	if err != nil {
		log.Printf("Error setting the read flag: ", err.Error())
	}
	log.Printf("Read flag set")
}

func updateInFile(inputFileName string) {
	println("updating .in file")
	inFilePath := "/docker/starlight/config_files_starlight/"
	inFileName := "grid_example.in"

	//fix this it;s shite
	if exists, direrr := exists(inFilePath + inFileName); exists == false && direrr == nil {
		println("Error: ", inFileName, " does not exist")
	}

	f, err := os.Open(inFilePath + inFileName)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	i := 0
	var newFile string
	for scanner.Scan() {
		i++
		if i == 16 {
			//not a great way to find th right line, look for better alternatives
			res := strings.Split(scanner.Text(), "  ")
			currentInputFileName := res[0]
			currentOutputFileName := res[5]
			println("input file = ", currentInputFileName, ", output file=", currentOutputFileName)

			//replace line 16

			res[0] = inputFileName
			//update this to append oput at end, before.txt
			res[5] = "output_" + inputFileName

			overwrite_string := strings.Join(res, "  ")

			println("New Line = ", overwrite_string)

			// write back to .in file
			newFile = newFile + overwrite_string + "\n"
		} else {
			newFile = newFile + scanner.Text() + "\n"
		}
		//println(i, " = ", scanner.Text()) // Println will add back the final '\n'
	}

	print("New File = ", newFile)

	filepath := "/starlight/"
	filename := "grid_example.in"

	//write new in file to /starlight/
	log.Printf("Writing updated .in file to %s", filepath+filename)
	werr := os.WriteFile(filepath+filename, []byte(newFile), 0644)

	if err != nil {
		println("Error ", werr.Error())
	}

	if err := scanner.Err(); err != nil {
		println(os.Stderr, "reading standard input:", err)
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
