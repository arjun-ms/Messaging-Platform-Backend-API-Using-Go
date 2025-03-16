package main // Defines this file as an executable program.

import (
	"context" 
	// "encoding/json" // Used to encode and decode JSON data.
	"fmt" // package for printing
	"log"  // Logs messages to the console with timestamps and severity levels.
	"time"
	"strings" // Provides utility functions for string manipulation.
	
	"github.com/jackc/pgx/v5" // PostgreSQL driver for Go
	"github.com/labstack/echo/v4" // Web framework for handling HTTP requests and building APIs.
	"github.com/redis/go-redis/v9" // Redis client for caching and real-time data handling.
	"github.com/google/uuid"
)

// Global connections
var (
	conn     *pgx.Conn  // db connection
	redisCli *redis.Client // redis connection
	ctx      = context.Background() // Global context used to manage request-scoped values, deadlines, and cancellation signals.
)

// Message struct for input (like a blueprint for objects)
type Message struct {
	// Field Type Tag
	MessageID    string    `json:"message_id"`
	SenderID     string    `json:"sender_id"`
	ReceiverID   string    `json:"receiver_id"`
	Content      string    `json:"content"`
	Timestamp    time.Time `json:"-"`    // Timestamp is skipped when converting to JSON because it's not needed in the response directly.
	TimestampStr string    `json:"timestamp"` // Instead, TimestampStr is used to convert it into a readable string format before sending it to the client.
	Read         bool      `json:"read"`
}


func main() {
	//! Connect to PostgreSQL
	var err error
	connString := "postgres://postgres:7591achu@localhost:5432/messaging" // Connection string to postgres
	conn, err = pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err) // If the connection fails, logs an error and exits.
	}
	defer conn.Close(context.Background()) // Closes the connection when the function exits.
	fmt.Println("Connected to PostgreSQL!")
	fmt.Println()

	//----------------------------------------------

	//! Connect to Redis
	// create a new redis client
	redisCli = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Default Redis port
		Password: "",               // No password
		DB:       0,                // Default DB
	})
	//  Sends a ping to Redis to check the connection
	_, err = redisCli.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v\n", err)
	}
	fmt.Println("Connected to Redis!")
	fmt.Println()

	//----------------------------------------------

	// Initialize Echo (for handling HTTP requests)
	e := echo.New() // sets up a lightweight HTTP server.
 
	// Define routes
	e.POST("/messages", sendMessage)
	e.GET("/messages", getMessages)
	e.PATCH("/messages/:id/read", markMessageAsRead)
	e.DELETE("/messages/:id", deleteMessage)


	// Start worker in a separate goroutine
	//! The go keyword starts the worker in a separate goroutine  (like a background thread).
	//! This allows the server and worker to run concurrently without blocking each other.
	go startWorker()

	// Start Echo server at 8080 or Change to any free port 
	e.Logger.Fatal(e.Start(":8080")) //  Fatal - If the server fails to start, logs an error and exits.
}

//! Handles sending a message (need asynchronous with a queue (redis Lists)) - working
// func sendMessage(c echo.Context) error {
// 	var msg Message

// 	// Bind request body(JSON input) to msg struct
// 	if err := c.Bind(&msg); err != nil {
// 		return c.JSON(400, map[string]string{"error": "Invalid request body"})
// 	}

// 	// Input validation (Ensures sender_id, receiver_id, and content are not empty.)
// 	if msg.SenderID == "" || msg.ReceiverID == "" || msg.Content == "" {
// 		return c.JSON(400, map[string]string{"error": "Sender ID, Receiver ID, and Content are required"})
// 	}

// 	// Convert message to JSON for pushing into Redis ( Converts struct to JSON for easy storage.)
// 	data, err := json.Marshal(msg)
// 	if err != nil {
// 		return c.JSON(500, map[string]string{"error": "Failed to process message"})
// 	}

// 	// Push to Redis queue
// 	// LPush → Adds the message to the Redis list message_queue.
// 	err = redisCli.LPush(context.Background(), "message_queue", data).Err()
// 	if err != nil {
// 		return c.JSON(500, map[string]string{"error": "Failed to queue message"})
// 	}

// 	return c.JSON(200, map[string]string{"status": "Message queued"})
// }


//! sendMessage (Using Redis Streams) - working
func sendMessage(c echo.Context) error {
	var msg Message
	if err := c.Bind(&msg); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid input"})
	}

	if msg.SenderID == "" || msg.ReceiverID == "" || msg.Content == "" {
		return c.JSON(400, map[string]string{"error": "Invalid message data"})
	}

	id := uuid.New().String()

	// Add to Redis Stream (instead of List)
	_, err := redisCli.XAdd(ctx, &redis.XAddArgs{
		Stream: "message_stream",
		Values: map[string]interface{}{
			"message_id":   id,
			"sender_id":    msg.SenderID,
			"receiver_id":  msg.ReceiverID,
			"content":      msg.Content,
			"timestamp":    time.Now().Format(time.RFC3339),
			"read":         false,
		},
	}).Result()

	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to add message to stream"})
	}
	
	log.Printf("Message queued with ID: %s\n", id)
	return c.JSON(200, map[string]string{"status": "Message queued"})
}


//! Handles retrieving conversation history between two users by using an SQL query - working
func getMessages(c echo.Context) error {
	log.Println("Starting to read messages from database...") // Debug log

	// Get query parameters
	user1 := c.QueryParam("user1")
	user2 := c.QueryParam("user2")

	// Validate query parameters
	if user1 == "" || user2 == "" {
		return c.JSON(400, map[string]string{"error": "user1 and user2 are required"})
	}

	// Define a SQL query to fetch messages between two users
	query := `
		SELECT message_id, sender_id, receiver_id, content, timestamp, read 
		FROM messages
		WHERE 
			(sender_id = $1 AND receiver_id = $2) OR 
			(sender_id = $2 AND receiver_id = $1)
		ORDER BY timestamp DESC
	`

	// Query on the Database to fetch the row
	rows, err := conn.Query(context.Background(), query, user1, user2)
	if err != nil {
		log.Printf("Failed to read messages: %v\n", err) // Debug log
		return c.JSON(500, map[string]string{"error": "Failed to fetch messages"})
	}
	defer rows.Close()

	// Fetch the messages and store them in a slice of Message structs.
	var messages []Message

	for rows.Next() {
		var msg Message

		// Scan the row into variables
		err := rows.Scan(&msg.MessageID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.Timestamp, &msg.Read)
		if err != nil {
			log.Printf("Failed to scan row: %v", err) // Debug log
			return c.JSON(500, map[string]string{"error": "Failed to read messages"})
		}

		// ✅ Convert Timestamp to string format for JSON
		msg.TimestampStr = msg.Timestamp.Format(time.RFC3339)

		messages = append(messages, msg)
		log.Printf("Fetched  Message: %+v", msg) // Debug log

	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err) // Debug log
		return c.JSON(500, map[string]string{"error": "Failed to process messages"})
	}

	// Return the fetched messages as JSON
	return c.JSON(200, messages)
}

//! Handles marking a message as read - working
func markMessageAsRead(c echo.Context) error {
	// Extract the message ID from the request URL
	messageID := c.Param("id")
	
	log.Printf("Marking message %s as read\n", messageID)

	// Validate input
	if messageID == "" {
		return c.JSON(400, map[string]string{"error": "Message ID is required"})
	}

	// Update the `read` status in the database
	query := `UPDATE messages SET read = TRUE WHERE message_id = $1`
	result, err := conn.Exec(context.Background(), query, messageID)
	if err != nil {
		log.Printf("Failed to update message status: %v\n", err)
		return c.JSON(500, map[string]string{"error": "Failed to update message status"})
	}

	// Check if any rows were affected (means message exists)
	if result.RowsAffected() == 0 {
		log.Printf("No message found with ID: %s\n", messageID)
		return c.JSON(404, map[string]string{"error": "Message not found"})
	}

	log.Printf("Message %s marked as read\n", messageID)
	return c.JSON(200, map[string]string{"status": "Message marked as read"})
}

// //! Handles deleting a message - working
func deleteMessage(c echo.Context) error {
	// fetch the id from the parameter passed during the request
	id := c.Param("id")

	// SQL query to delete the message by ID
	query := `DELETE FROM messages WHERE message_id = $1` // $1 is a positional placeholder used in PostgreSQL for parameterized queries.
	result, err := conn.Exec(context.Background(), query, id) //  binds the id value to $1 safely (prevents SQL Injection)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		return c.JSON(500, map[string]string{"error": "Failed to delete message"})
	}

	// Check if any rows were affected
	if result.RowsAffected() == 0 {
		return c.JSON(404, map[string]string{"error": "Message not found"})
	}

	return c.JSON(200, map[string]string{"status": "Message deleted"})
}


//! Process messages from the Redis queue and store them in PostgreSQL (for redis List)
// func startWorker() {
// 	for {
// 		// BRPop → Blocking pop to wait for new messages in the queue.
// 		data, err := redisCli.BRPop(context.Background(), 0, "message_queue").Result()
// 		if err != nil {
// 			log.Printf("Failed to read from queue: %v\n", err)
// 			continue
// 		}

// 		if len(data) < 2 {
// 			continue
// 		}

// 		var msg Message
// 		err = json.Unmarshal([]byte(data[1]), &msg)
// 		if err != nil {
// 			log.Printf("Failed to parse message: %v\n", err)
// 			continue
// 		}

// 		// Insert into PostgreSQL
// 		query := `INSERT INTO messages (sender_id, receiver_id, content, timestamp, read)
// 		          VALUES ($1, $2, $3, $4, $5)`
// 		_, err = conn.Exec(context.Background(), query, msg.SenderID, msg.ReceiverID, msg.Content, time.Now(), false)
// 		if err != nil {
// 			log.Printf("Failed to insert message into database: %v\n", err)
// 			continue
// 		}

// 		fmt.Println("✅ Message stored in database:", msg)
// 	}
// }

// !-----------------------------------------------------


//! Worker for Redis Streams
func startWorker() {

	log.Println("Starting Redis stream worker...")

	// Create Consumer Group (if not exists)
	_, err := redisCli.XGroupCreateMkStream(ctx, "message_stream", "message_group", "$").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Fatalf("Failed to create consumer group: %v", err)
	}

	for {
		// Read from the stream using a consumer group
		streams, err := redisCli.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "message_group",
			Consumer: "worker-1",
			Streams:  []string{"message_stream", ">"},
			Block:    0,
			Count:    1,
		}).Result()

		if err != nil {
			log.Printf("Failed to read from stream: %v", err)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				// This point is right after messages are fetched from Redis but before they're inserted into PostgreSQL.
				fmt.Println("\nReceived message from stream: ", message)

				messageID := message.ID
				senderID := message.Values["sender_id"].(string)
				receiverID := message.Values["receiver_id"].(string)
				content := message.Values["content"].(string)
				timestamp := message.Values["timestamp"].(string)

				// Insert into PostgreSQL
				_, err := conn.Exec(context.Background(),
					"INSERT INTO messages (message_id, sender_id, receiver_id, content, timestamp, read) VALUES ($1, $2, $3, $4, $5, $6)",
					messageID, senderID, receiverID, content, timestamp, false)

				if err != nil {
					log.Printf("Failed to insert message: %v", err)
					continue
				}else{
					log.Printf("✅ Message inserted into DB with ID: %s\n", messageID)
				}

				// Acknowledge the message after processing
				_, err = redisCli.XAck(ctx, "message_stream", "message_group", messageID).Result()
				if err != nil {
					log.Printf("Failed to ACK message: %v", err)
				}else {
					log.Printf("✅ Message ACKed: %s\n", messageID)
				}
			}
		}
	}
}
