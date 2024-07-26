package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	//"time"
	"math/rand"
	"strconv"

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

	println("user ", username)
	println("pass ", password)
	println("host ", host)
	println("port ", port)
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
		newInputFileName := ""
		for {

			//if checkReadFlag() == false {

			//	log.Println("Read Flag = ", checkReadFlag())
			d := <-msgs

			//for d := range msgs {
			log.Printf(" -------------------------- Received a message: %s -------------------------------", d.Headers["filename"])

			//d.Ack(false)

			//filename := string(d.Headers["filename"])

			if filename, ok := d.Headers["filename"].(string); ok {
				writeMessageToFile(d.Body, filename)
				newInputFileName = updateInFile(filename)
				//setReadFlag()

				updateToProcessList(newInputFileName)

				outputPath := "/starlight/data/output"
				//check if output dir exists, if not create as not sure if Starlight will crete it
				if exists, direrr := exists(outputPath); exists == false && direrr == nil {
					println("Output directory does not exist, making it", outputPath)
					err := os.Mkdir(outputPath, 0700)
					if err != nil {
						println("ERROR: ", err)
					}
				}

				d.Ack(false)
			} else {
				log.Println("Error ", ok)
			}
			//}
		}
		log.Printf("Processed Messages")
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

func writeMessageToFile(body []byte, filename string) {
	//TODO - review this, at the moment all data on a shared drive so this is actually unneccessary

	println("Writing message to data file")
	filepath := os.Getenv("DATA_FILE_PATH")

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
	DATA_FILE_FLAG := os.Getenv("DATA_FILE_FLAG")
	if _, err := os.Stat(DATA_FILE_FLAG); err == nil {
		//log.Println("File "+fi.Name())
		return true
	}
	return false
}

func updateToProcessList(inFileName string) {
	log.Printf("Adding new .in file to ProcessList %s", inFileName)
	PROCESS_LIST := os.Getenv("PROCESS_LIST")

	//fix this it;s shite
	if exists, direrr := exists(PROCESS_LIST); exists == false && direrr == nil {
		println("Error: ", PROCESS_LIST, " does not exist, creating it")
		//creating file
		touchFile(PROCESS_LIST)
		println(PROCESS_LIST, " has been created")
	}

	f, err := os.OpenFile(PROCESS_LIST, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer f.Close()

	if err != nil {
		println("Error: could not open", PROCESS_LIST)
		panic(err)
	}

	//add new .in file to bottom of list

	if _, err = f.WriteString(inFileName + "\n"); err != nil {
		panic(err)
	}
	println(inFileName, " has been added")
}

func setReadFlag() {
	// creates the start_starlight file
	log.Printf("Setting the read flag")
	DATA_FILE_FLAG := os.Getenv("DATA_FILE_FLAG")

	d1 := []byte("start")
	err := os.WriteFile(DATA_FILE_FLAG, d1, 0666)
	if err != nil {
		log.Printf("Error setting the read flag: ", err.Error())
	}
	log.Printf("Read flag set")
}

func updateInFile(inputFileName string) string {
	println("updating .in file")
	templateInFilePath := "/docker/starlight/config_files_starlight/"
	templateInFileName := "grid_example.in"
	inFileOutputPath := "/starlight/runtime/infiles/"
	newInFileName := "grid_example" + strconv.Itoa(rand.Int()) + ".in"

	//fix this it;s shite
	if exists, direrr := exists(templateInFilePath + templateInFileName); exists == false && direrr == nil {
		println("Error: ", templateInFileName, " does not exist")
	}

	f, err := os.Open(templateInFilePath + templateInFileName)
	defer f.Close()
	if err != nil {
		println("Error opening file")
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
	//filename := "grid_example.in"

	//write new in file to /starlight/
	log.Printf("Writing updated .in file to %s", inFileOutputPath+newInFileName)
	werr := os.WriteFile(inFileOutputPath+newInFileName, []byte(newFile), 0644)

	if err != nil {
		println("Error ", werr.Error())
	}

	if err := scanner.Err(); err != nil {
		println(os.Stderr, "reading standard input:", err)
	}

	return newInFileName
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
