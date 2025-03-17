# Messaging Platform Backend API Using Go

Welcome to the **Messaging Platform Backend API** project! This backend service facilitates real-time messaging capabilities, built using the Go programming language.

## Table of Contents

- [Demo](#demo)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [Usage](#usage)
- [API Documentation](#api-documentation)

## Demo

[Watch the Demo](https://www.loom.com/embed/aa24d1f6a65d4c01ba9b8c37bc79cd1b?sid=cd8af047-168d-4a18-8a99-67cb45d852a1)


## Features

- **Real-time Messaging**: Supports instant message delivery between users.
- **Message Persistence**: Stores message history for retrieval.

## Getting Started

Follow these instructions to set up and run the project on your local machine.

### Prerequisites

- **Go**: Ensure you have Go installed. You can download it from the [official website](https://golang.org/dl/).
- **PostgreSQL**: This project uses PostgreSQL as the primary database. Install it from [here](https://www.postgresql.org/download/).
- **Redis**: Used for caching and real-time message brokering. Installation instructions are available [here](https://redis.io/download).

### Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/arjun-ms/Messaging-Platform-Backend-API-Using-Go.git
   cd Messaging-Platform-Backend-API-Using-Go
   ```

2. **Install Dependencies**:

   Ensure all Go dependencies are installed by running:

   ```bash
   go mod tidy
   ```

3. **Set Up the Database**:

   - Create a PostgreSQL database named `messaging_platform`.
   - Create a table called `Message`
     | Field | Type | Description |
     |-------|------|-------------|
     | message_id | string | Unique ID for the message |
     | sender_id | string | ID of the sender |
     | receiver_id | string | ID of the receiver |
     | content | string | Message content |
     | timestamp | string | Message timestamp (RFC3339) |
     | read | boolean | Message read status |
     | status | string | Message status (sent, delivered, read) |

## Usage

To start the server, execute:

```bash
go run main.go
```

The API will be accessible at `http://localhost:8080`.

## API Documentation

For detailed API endpoints and request/response formats, refer to the [DOCUMENTATION.md](DOCUMENTATION.md) file.
