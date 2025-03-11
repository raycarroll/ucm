package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DataFile struct {
	Name    string
	Content string
}

type Event struct {
	Files []DataFile
}

type Producer struct {
	batchSize  int
	batch      []DataFile
	eventQueue chan Event
	inputDir   string
	outputDir  string
}

func NewProducer(batchSize int, inputDir, outputDir string, eventQueue chan Event) *Producer {
	return &Producer{
		batchSize:  batchSize,
		batch:      make([]DataFile, 0, batchSize),
		eventQueue: eventQueue,
		inputDir:   inputDir,
		outputDir:  outputDir,
	}
}

func (p *Producer) AddFile(file DataFile) {
	p.batch = append(p.batch, file)
	if len(p.batch) >= p.batchSize {
		p.sendBatch()
	}
}

func (p *Producer) sendBatch() {
	if len(p.batch) > 0 {
		// Update the .in file before sending the batch
		inFileName, content := p.updateInFile()
		log.Printf(inFileName)
		log.Printf(content)
		if inFileName != "" && content != "" {
			p.batch = append(p.batch, DataFile{Name: inFileName, Content: content})
		}

		event := Event{Files: p.batch}
		p.eventQueue <- event
		p.removeInFileFromBatch()
		p.moveProcessedFiles()
		p.batch = make([]DataFile, 0, p.batchSize)
	}
}

func (p *Producer) moveProcessedFiles() {
	for _, file := range p.batch {
		sourcePath := filepath.Join(p.inputDir, file.Name)
		destPath := filepath.Join(p.outputDir, file.Name)
		err := MoveFile(sourcePath, destPath)
		if err != nil {
			fmt.Printf("Error moving file %s: %v\n", file.Name, err)
		}
	}
}

func (p *Producer) ReadFiles() {
	files, err := os.ReadDir(p.inputDir)
	failOnError(err, "Failed reading input directory")
	for _, file := range files {
		if !file.IsDir() {
			content, err := os.ReadFile(filepath.Join(p.inputDir, file.Name()))
			if err != nil {
				log.Printf("Error reading file %s: %v\n", file.Name(), err)
				continue
			}
			p.AddFile(DataFile{Name: file.Name(), Content: string(content)})
		}
	}
}

func (p *Producer) updateInFile() (string, string) {
	println("updating .in file")
	templateInFilePath := os.Getenv("TEMPLATE_IN_FILE_PATH")
	inFileOutputPath := os.Getenv("IN_FILE_OUTPUT_PATH")
	/*templateInFilePath := "/docker/starlight/config_files_starlight/grid_example.in"
	inFileOutputPath := "/starlight/runtime/infiles/" */
	newInFileName := fmt.Sprintf("grid_example_%d.in", rand.Intn(100))

	// Check if the template .in file exists
	if exists, _ := exists(templateInFilePath); !exists {
		println("Error: file does not exist")
		return "", ""
	}

	f, err := os.Open(templateInFilePath)
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
			// Replace the input file name in the .in file
			res := strings.Split(scanner.Text(), "  ")
			for j := 0; j < len(p.batch); j++ {
				res[0] = p.batch[j].Name

				// Get kinematic values for the current file
				kinematicValues, err := p.getKinematicValues(p.batch[j].Name)
				if err != nil {
					log.Printf("Error getting kinematic values for file %s: %v", p.batch[j].Name, err)
					continue
				}
				res[4] = kinematicValues // Update the 4th and 5th parameters with Velocity and Sigma
				res[5] = "output_" + p.batch[j].Name
				overwrite_string := strings.Join(res, "  ")
				newFile = newFile + overwrite_string + "\n"
			}
		} else {
			newFile = newFile + scanner.Text() + "\n"
		}
	}

	// Write the updated .in file to the output directory
	log.Printf("Writing updated .in file to %s", inFileOutputPath+newInFileName)
	err = os.WriteFile(inFileOutputPath+newInFileName, []byte(newFile), 0644)
	if err != nil {
		println("Error writing .in file: ", err.Error())
		return "", ""
	}

	// Read the content of the new .in file
	content, err := os.ReadFile(inFileOutputPath + newInFileName)
	if err != nil {
		println("Error reading the newly created .in file:", err.Error())
		return "", ""
	}

	return newInFileName, string(content)
}
func (p *Producer) getKinematicValues(fileName string) (string, error) {
	kinematicFilePath := "./data/kinematic_information_file_NGC7025_LR-V.txt"
	file, err := os.Open(kinematicFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open kinematic file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if fields[0] == fileName {
			velocity := fields[1]
			sigma := fields[3]
			log.Printf("%s %s", velocity, sigma)
			return fmt.Sprintf("%s %s", velocity, sigma), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading kinematic file: %v", err)
	}

	return "", fmt.Errorf("file %s not found in kinematic information", fileName)
}

func (p *Producer) removeInFileFromBatch() {
	log.Printf("Removing .in file from batch")
	filteredBatch := make([]DataFile, 0, len(p.batch))

	for _, file := range p.batch {
		if !strings.HasSuffix(file.Name, ".in") {
			filteredBatch = append(filteredBatch, file)
		} else {
			inFilePath := filepath.Join("/starlight/runtime/infiles/", file.Name)
			err := os.Remove(inFilePath)
			if err != nil {
				log.Printf("Error removing .in file %s: %v\n", inFilePath, err)
			} else {
				log.Printf("Successfully removed .in file: %s\n", inFilePath)
			}
		}

	}

	p.batch = filteredBatch
}

func mainRun() {
	eventQueue := make(chan Event, 10)

	inputDir := os.Getenv("INPUT_DIR")
	outputDir := os.Getenv("OUTPUT_DIR")

	if inputDir == "" || outputDir == "" {
		log.Fatal("INPUT_DIR  and OUTPUT_DIR environment variables is required")
	}

	batchSize, err := strconv.Atoi("5")
	if err != nil {
		log.Fatal("Failed to convert BATCH_SIZE to an integer")
	}
	producer := NewProducer(batchSize, inputDir, outputDir, eventQueue)
	go func() {
		for event := range eventQueue {
			log.Printf("Sent event with %d files\n", len(event.Files))
			send(event)
		}
	}()
	for {
		producer.ReadFiles()
		producer.sendBatch()
		time.Sleep(10 * time.Second)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}

func MoveFile(source, destination string) error {
	return os.Rename(source, destination)
}

func send(event Event) {
	username := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)

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

	for _, f := range event.Files {
		body := f.Content
		headers := make(amqp.Table)
		headers["filename"] = f.Name

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

		// Safely log the first 10 characters of the body
		if len(body) >= 10 {
			log.Printf(" [x] Sent %s\n", body[0:10])
		} else {
			log.Printf(" [x] Sent %s\n", body)
		}
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
