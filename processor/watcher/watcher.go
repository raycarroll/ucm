package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DataFile struct {
	Name    string
	Content string
}

type Event struct {
	Files []DataFile
}

type Watcher struct {
	batchSize  int
	batch      []DataFile
	eventQueue chan Event
	dirToWatch string
}

func NewWatcher(batchSize int, dirToWatch string, eventQueue chan Event) *Watcher {
	return &Watcher{
		batchSize:  batchSize,
		batch:      make([]DataFile, 0, batchSize),
		eventQueue: eventQueue,
		dirToWatch: dirToWatch,
	}
}

// Add files to batch
func (w *Watcher) addFile(file DataFile) {
	w.batch = append(w.batch, file)
	if len(w.batch) >= w.batchSize {
		w.sendBatch()
	}
}

func (w *Watcher) sendBatch() {
	if len(w.batch) > 0 {
		event := Event{Files: w.batch}
		w.eventQueue <- event
		w.deleteProcessedFiles()
		w.batch = make([]DataFile, 0, w.batchSize)
	}
}

// Remove processed files from the watched directory
func (w *Watcher) deleteProcessedFiles() {
	for _, file := range w.batch {
		err := os.Remove(file.Name)
		if err != nil {
			failOnError(err, fmt.Sprintf("Error deleting file: %s", file.Name))
		}
	}
}

func (w *Watcher) readFiles(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", filename, err)
		return
	}
	w.addFile(DataFile{Name: filename, Content: string(content)})
}

// onFileWrite is called when a file is created in the watched directory
func (w *Watcher) onFileWrite(filename string) {
	fmt.Printf("File written: %s\n", filename)
	w.readFiles(filename)
	w.sendBatch()
}

func (w *Watcher) watcher() {
	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	// Ensure the directory exists
	if _, err := os.Stat(w.dirToWatch); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s\n", w.dirToWatch)
	}

	// Add the directory to the watcher
	err = watcher.Add(w.dirToWatch)
	if err != nil {
		log.Fatal("Error adding directory to watcher:", err)
	}
	log.Printf("Watching directory: %s\n", w.dirToWatch)

	// Listen for events
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Check if the event is a file creation
				if event.Op&fsnotify.Create == fsnotify.Create {
					w.onFileWrite(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()
	<-done
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
		"watcher", // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
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
		log.Printf(" [x] Sent %s\n", body[0:10])
	}
}

func processQueue(eventQueue chan Event) {
	for event := range eventQueue {
		fmt.Printf("Processing event with %d files\n", len(event.Files))
		send(event)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}

func main() {
	eventQueue := make(chan Event, 10)
	// Directory to watch
	dirToWatch := os.Getenv("INPUT_DIR")
	if dirToWatch == "" {
		log.Fatal("INPUT_DIR environment variable is required")
	}

	batchSize, err := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	if err != nil {
		log.Fatal("Failed to convert BATCH_SIZE to an integer")
	}

	w := NewWatcher(batchSize, dirToWatch, eventQueue)
	go processQueue(eventQueue)
	w.watcher()
}
