package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type FileData struct {
	Name    string `json:"Name"`
	Content string `json:"Content"`
}

type MessageBody struct {
	Files []FileData `json:"Files"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("------------------ Starting receive() ---------------------")
	// Initialize directories first
	if err := ensureDirectoriesExist(); err != nil {
		log.Fatalf("Directory initialization failed: %v", err)
	}

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

	// Declare queues for each application
	queues := []string{"starlight", "ppfx", "steckmap"}
	for _, queue := range queues {
		_, err = ch.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		failOnError(err, fmt.Sprintf("Failed to declare queue: %s", queue))
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	// Process queues continuously
	for {
		for _, queue := range queues {
			processQueue(ch, queue)
		}
		time.Sleep(1 * time.Second) // Brief pause between queue checks
	}
}
func ensureDirectoriesExist() error {
	requiredDirs := []string{
		os.Getenv("INPUT_DIR_STARLIGHT"),
		os.Getenv("OUTPUT_DIR_STARLIGHT"),
		os.Getenv("INPUT_DIR_PPFX"),
		os.Getenv("OUTPUT_DIR_PPFX"),
		os.Getenv("INPUT_DIR_STECKMAP"),
		os.Getenv("OUTPUT_DIR_STECKMAP"),
		os.Getenv("IN_FILE_PATH"),
	}

	// Create all required directories
	for _, dir := range requiredDirs {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
		log.Printf("Verified directory: %s", dir)
	}

	// Create process list file if specified
	processListPath := os.Getenv("PROCESS_LIST")
	log.Printf(processListPath)
	if processListPath != "" {
		// Ensure parent directory exists
		parentDir := filepath.Dir(processListPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for process list: %v", err)
		}

		// Create empty file if it doesn't exist
		if _, err := os.Stat(processListPath); os.IsNotExist(err) {
			file, err := os.Create(processListPath)
			if err != nil {
				return fmt.Errorf("failed to create process list file: %v", err)
			}
			file.Close()
			log.Printf("Created process list file: %s", processListPath)
		} else if err != nil {
			return fmt.Errorf("failed to check process list file: %v", err)
		}
	}

	return nil
}

func processQueue(ch *amqp.Channel, queue string) {
	// Get initial queue info
	queueInfo, err := ch.QueueInspect(queue)
	if err != nil {
		log.Printf("QUEUE ERROR: Failed to inspect queue %s: %v", queue, err)
		return
	}

	if queueInfo.Messages == 0 {
		return // No messages to process
	}

	log.Printf("\n==============================================")
	log.Printf("PROCESSING QUEUE: %s (%d messages)", queue, queueInfo.Messages)
	log.Printf("==============================================")

	// Generate a unique consumer tag
	consumerTag := fmt.Sprintf("consumer-%s-%d", queue, time.Now().UnixNano())

	// Start consuming
	msgs, err := ch.Consume(
		queue,       // queue
		consumerTag, // consumer tag
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.Printf("CONSUME ERROR: Failed to register consumer for queue %s: %v", queue, err)
		return
	}

	defer func() {
		if err := ch.Cancel(consumerTag, false); err != nil {
			log.Printf("WARNING: Failed to cancel consumer %s: %v", consumerTag, err)
		}
	}()

	// Process messages for a limited time
	timeout := time.After(5 * time.Second)

	for {
		select {
		case d, ok := <-msgs:
			if !ok {
				return // Channel closed
			}
			processMessage(ch, queue, d)
		case <-timeout:
			return
		}
	}
}

func processMessage(ch *amqp.Channel, queue string, d amqp.Delivery) {
	processStart := time.Now()
	batchID := fmt.Sprintf("%s-%d", queue, d.DeliveryTag)

	log.Printf("\n■■■ BATCH START [%s] ■■■", batchID)
	log.Printf("│ Queue:      %s", queue)
	log.Printf("│ DeliveryTag: %d", d.DeliveryTag)
	log.Printf("│ Timestamp:   %s", d.Timestamp)

	// Log all headers
	if len(d.Headers) > 0 {
		headers, _ := json.Marshal(d.Headers)
		log.Printf("│ Headers:    %s", headers)
	}

	// Get batch information from headers
	batchSize, ok := d.Headers["batch_size"].(int32)
	if !ok {
		log.Printf("│ ERROR: 'batch_size' header missing or invalid")
		d.Nack(false, true)
		log.Printf("■■■ BATCH ERROR [%s] - Message requeued ■■■", batchID)
		return
	}

	filenamesHeader, ok := d.Headers["filenames"].(string)
	if !ok {
		log.Printf("│ ERROR: 'filenames' header missing or invalid")
		d.Nack(false, true)
		log.Printf("■■■ BATCH ERROR [%s] - Message requeued ■■■", batchID)
		return
	}

	filenames := strings.Split(filenamesHeader, ",")
	if len(filenames) != int(batchSize) {
		log.Printf("│ ERROR: Filenames count doesn't match batch_size")
		d.Nack(false, true)
		log.Printf("■■■ BATCH ERROR [%s] - Message requeued ■■■", batchID)
		return
	}

	log.Printf("│ Processing batch of %d files:", batchSize)
	for i, filename := range filenames {
		log.Printf("│ %d. %s", i+1, filename)
	}

	var outputPath string
	switch queue {
	case "starlight":
		outputPath = os.Getenv("INPUT_DIR_STARLIGHT")
	case "ppfx":
		outputPath = os.Getenv("INPUT_DIR_PPFX")
	case "steckmap":
		outputPath = os.Getenv("INPUT_DIR_STECKMAP")
	default:
		log.Printf("│ ERROR: Unknown queue: %s", queue)
		d.Nack(false, true)
		log.Printf("■■■ BATCH ERROR [%s] - Message requeued ■■■", batchID)
		return
	}

	if outputPath == "" {
		log.Printf("│ ERROR: Output directory not configured for queue")
		d.Nack(false, true)
		log.Printf("■■■ BATCH ERROR [%s] - Message requeued ■■■", batchID)
		return
	}

	// Parse message body
	var msgBody MessageBody
	err := json.Unmarshal(d.Body, &msgBody)
	if err != nil {
		log.Printf("│ ERROR parsing message body: %v", err)
		d.Nack(false, true)
		log.Printf("■■■ BATCH ERROR [%s] - Message requeued ■■■", batchID)
		return
	}

	if len(msgBody.Files) != int(batchSize) {
		log.Printf("│ ERROR: Files count in body doesn't match batch_size")
		d.Nack(false, true)
		log.Printf("■■■ BATCH ERROR [%s] - Message requeued ■■■", batchID)
		return
	}

	// Ensure output directory exists
	if exists, _ := exists(outputPath); !exists {
		err := os.Mkdir(outputPath, 0700)
		if err != nil {
			log.Printf("│ ERROR creating directory: %v", err)
			d.Nack(false, true)
			log.Printf("■■■ BATCH ERROR [%s] - Message requeued ■■■", batchID)
			return
		}
		log.Printf("│ Created output directory: %s", outputPath)
	}

	successCount := 0
	for _, file := range msgBody.Files {
		if strings.HasSuffix(file.Name, ".in") {
			updateToProcessList(file.Name, []byte(file.Content))
			log.Printf("│ ✓ Processed .in file: %s", file.Name)
			successCount++
			continue
		}

		filePath := filepath.Join(outputPath, file.Name)
		err := os.WriteFile(filePath, []byte(file.Content), 0644)
		if err != nil {
			log.Printf("│ ✗ Error writing file %s: %v", file.Name, err)
		} else {
			log.Printf("│ ✓ Wrote file: %s", file.Name)
			successCount++
		}
	}

	if successCount == int(batchSize) {
		d.Ack(false)
		log.Printf("│ ✔ Successfully processed all %d files", batchSize)
	} else {
		log.Printf("│ ⚠ Processed %d/%d files successfully", successCount, batchSize)
		d.Nack(false, true)
	}

	log.Printf("│ Duration: %s", time.Since(processStart))
	log.Printf("■■■ BATCH COMPLETE [%s] ■■■", batchID)
}

func updateToProcessList(inFileName string, fileContent []byte) {
	PROCESS_LIST := os.Getenv("PROCESS_LIST")
	InFilePath := os.Getenv("IN_FILE_PATH")

	if err := touchFile(PROCESS_LIST); err != nil {
		log.Printf("│ ✗ Error creating process list: %v", err)
		return
	}

	specialFilePath := filepath.Join(InFilePath, inFileName)
	err := os.WriteFile(specialFilePath, fileContent, 0644)
	if err != nil {
		log.Printf("│ ✗ Error writing .in file: %v", err)
		return
	}

	f, err := os.OpenFile(PROCESS_LIST, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Printf("│ ✗ Error opening process list: %v", err)
		return
	}
	defer f.Close()

	if _, err = f.WriteString(inFileName + "\n"); err != nil {
		log.Printf("│ ✗ Error updating process list: %v", err)
	}
}

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
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
