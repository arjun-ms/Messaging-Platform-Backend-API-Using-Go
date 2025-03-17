# Messaging Platform API Documentation

## Overview
This API allows users to send, receive, and manage messages between users in a messaging platform. 
It is built using **Go**, **Echo** (for HTTP handling), **PostgreSQL** (for message storage), and **Redis Streams** (for message queuing).

## Base URL
```
http://localhost:8080
```

## Endpoints

### 1. **Get Messages**
- **Endpoint:** `/messages`
- **Method:** `GET`
- **Description:** Retrieves conversation history between two users.
- **Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| user1 | string | Yes | User ID/Name of the first participant |
| user2 | string | Yes | User ID/Name of the second participant |

- **Example Request:**
```
GET /messages?user1=123&user2=456
```

- **Example Response:**
```json
[
  {
    "message_id": "abc-123",
    "sender_id": "123",
    "receiver_id": "456",
    "content": "Hello!",
    "timestamp": "2025-03-15T12:00:00Z",
    "read": false,
    "status": "sent"
  },
  .
  .
  .
]
```

- **Possible Status Codes:**
  - `200 OK` – Successfully retrieved messages.
  - `400 Bad Request` – Missing query parameters.
  - `500 Internal Server Error` – Error while fetching messages.

---

### 2. **Send Message**
- **Endpoint:** `/messages`
- **Method:** `POST`
- **Description:** Sends a new message using Redis Streams.
- **Request Body:**
```json
{
  "sender_id": "user1",
  "receiver_id": "user2",
  "content": "Hello!"
}
```

- **Example Response:**
```json
{
  "status": "Message queued"
}
```

- **Possible Status Codes:**
  - `200 OK` – Message queued successfully.
  - `400 Bad Request` – Invalid input.
  - `500 Internal Server Error` – Error adding message to Redis stream.

---

### 3. **Mark Message as Read**
- **Endpoint:** `/messages/:id/read`
- **Method:** `PATCH`
- **Description:** Marks a message as read.
- **Example Request:**
```
PATCH /messages/abc-123/read
```

- **Example Response:**
```json
{
  "status": "Message marked as read"
}
```

- **Possible Status Codes:**
  - `200 OK` – Message status updated.
  - `400 Bad Request` – Missing or invalid ID.
  - `404 Not Found` – Message not found.
  - `500 Internal Server Error` – Error updating message.

---

### 4. **Mark Message as Delivered**
- **Endpoint:** `/messages/:id/delivered`
- **Method:** `PUT`
- **Description:** Marks a message as delivered.
- **Example Request:**
```
PUT /messages/abc-123/delivered
```

- **Example Response:**
```json
{
  "message": "Message status updated to delivered"
}
```

- **Possible Status Codes:**
  - `200 OK` – Message status updated.
  - `500 Internal Server Error` – Error updating message.

---

### 5. **Delete Message**
- **Endpoint:** `/messages/:id`
- **Method:** `DELETE`
- **Description:** Deletes a message by ID.
- **Example Request:**
```
DELETE /messages/abc-123
```

- **Example Response:**
```json
{
  "status": "Message deleted"
}
```

- **Possible Status Codes:**
  - `200 OK` – Message deleted.
  - `404 Not Found` – Message not found.
  - `500 Internal Server Error` – Error deleting message.

---

### 6. **Stop Redis Worker**
- **Endpoint:** `/stop-redis`
- **Method:** `POST`
- **Description:** Stops the Redis stream worker.

- **Example Request:**
```
POST /stop-redis
```

- **Example Response:**
```json
{
  "status": "Redis worker stopped"
}
```

- **Possible Status Codes:**
  - `200 OK` – Worker stopped successfully.

<br>

---
---

<br>

## Data Model
### Message
| Field | Type | Description |
|-------|------|-------------|
| message_id | string | Unique ID for the message |
| sender_id | string | ID of the sender |
| receiver_id | string | ID of the receiver |
| content | string | Message content |
| timestamp | string | Message timestamp (RFC3339) |
| read | boolean | Message read status |
| status | string | Message status (sent, delivered, read) |

---

## Technologies Used
- **Go** – Programming language
- **Echo** – Web framework
- **PostgreSQL** – Database
- **Redis** – Stream-based message queue
- **UUID** – Unique ID generator

---

## Setup Instructions
1. Clone the repository:
    ```
    git clone https://github.com/arjun-ms/go-messaging-platform.git
    ```
2. Navigate to the project directory:
    ```
    cd messaging-platform
    ```
3. Start PostgreSQL and Redis servers.
4. Set up environment variables for database connection:
    ```
    export DATABASE_URL="postgres://username:password@localhost:5432/messaging"
    ```
5. Run the server:
    ```
    go run main.go
    ```

