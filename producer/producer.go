package main

import (
	"bufio"
	"context"
	"encoding/json"
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

func (p *Producer) AddFile(file DataFile, appName string) {
	p.batch = append(p.batch, file)
	if len(p.batch) >= p.batchSize {
		p.sendBatch(appName)
	}
}

func (p *Producer) sendBatch(appName string) {
	if len(p.batch) > 0 {
		// Update the .in file before sending the batch
		if appName == "starlight" {
			inFileName, content := p.updateInFile()
			//log.Printf(inFileName)
			//log.Printf(content)
			if inFileName != "" && content != "" {
				p.batch = append(p.batch, DataFile{Name: inFileName, Content: content})
			}
		}
		event := Event{Files: p.batch}
		p.eventQueue <- event

		if appName == "starlight" {
			p.removeInFileFromBatch()
		}

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

func (p *Producer) ReadFiles(appName string) {
	files, err := os.ReadDir(p.inputDir)
	failOnError(err, "Failed reading input directory")
	for _, file := range files {
		if !file.IsDir() {
			content, err := os.ReadFile(filepath.Join(p.inputDir, file.Name()))
			if err != nil {
				log.Printf("Error reading file %s: %v\n", file.Name(), err)
				continue
			}
			p.AddFile(DataFile{Name: file.Name(), Content: string(content)}, appName)
		}
	}
}

func (p *Producer) updateInFile() (string, string) {
	//println("updating .in file")
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
	//log.Printf("Writing updated .in file to %s", inFileOutputPath+newInFileName)
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

			return fmt.Sprintf("%s %s", velocity, sigma), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading kinematic file: %v", err)
	}

	return "", fmt.Errorf("file %s not found in kinematic information", fileName)
}
func (p *Producer) removeInFileFromBatch() {
	//log.Printf("Removing .in file from batch")
	filteredBatch := make([]DataFile, 0, len(p.batch))

	for _, file := range p.batch {
		if !strings.HasSuffix(file.Name, ".in") {
			filteredBatch = append(filteredBatch, file)
		} else {
			inFilePath := filepath.Join("/processing_data/starlight/runtime/infiles/", file.Name)
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
	// Run all three applications concurrently
	for {
		ppfxApp()
		starlightApp()
		steckmapApp()
		log.Println("Checking for new files...")
		time.Sleep(10 * time.Second)
	}
}

func createEvent(inputDir string, outputDir string, appName string) {
	eventQueue := make(chan Event, 10)

	if inputDir == "" || outputDir == "" {
		log.Fatal("INPUT_DIR and OUTPUT_DIR environment variables are required")
	}

	batchSize, err := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	if err != nil {
		log.Fatal("Failed to convert BATCH_SIZE to an integer")
	}

	producer := NewProducer(batchSize, inputDir, outputDir, eventQueue)

	go func() {
		for event := range eventQueue {
			log.Printf("Sent event with %d files\n", len(event.Files))
			send(event, appName)
		}
	}()

	producer.ReadFiles(appName)
	producer.sendBatch(appName)

}

// function to check directories and process files
func getDirectoriesWithFiles(inputDirEnv string, outputDirEnv string, appName string) {
	inputDir := os.Getenv(inputDirEnv)
	outputDir := os.Getenv(outputDirEnv)

	if inputDir == "" || outputDir == "" {
		log.Printf("%s directories not set\n", appName)
		return
	}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Printf("Error reading %s input directory: %v\n", appName, err)
		return
	}

	if len(files) > 0 {
		log.Printf("Processing %s files...\n", appName)
		createEvent(inputDir, outputDir, appName)
	} else {
		log.Printf("No files found in %s directories\n", appName)
	}
}
func starlightApp() {
	getDirectoriesWithFiles("INPUT_DIR_Starlight", "OUTPUT_DIR_Starlight", "starlight")
}
func ppfxApp() {
	getDirectoriesWithFiles("INPUT_DIR_PPFX", "OUTPUT_DIR_PPFX", "ppfx")
}
func steckmapApp() {
	getDirectoriesWithFiles("INPUT_DIR_Steckmap", "OUTPUT_DIR_Steckmap", "steckmap")
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}
func MoveFile(source, destination string) error {
	return os.Rename(source, destination)
}

func send(event Event, appName string) {
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
		appName, // dynamic queue name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Serialize the entire event (batch of files) to JSON
	eventJSON, err := json.Marshal(event)
	failOnError(err, "Failed to marshal event to JSON")

	headers := make(amqp.Table)
	headers["batch_size"] = len(event.Files)

	// Include all filenames in headers
	var filenames []string
	for _, f := range event.Files {
		filenames = append(filenames, f.Name)
	}
	headers["filenames"] = strings.Join(filenames, ",")

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        eventJSON,
			Headers:     headers,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent batch with %d files\n", len(event.Files))
	log.Printf("     Files: %s\n", strings.Join(filenames, ", "))
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
