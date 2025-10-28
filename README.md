# TinyLink Go URL Shortener Service (MongoDB)

This is a RESTful API service for shortening URLs, built with Go. It uses a Clean Architecture approach to separate concerns and MongoDB for persistent storage.

<img width="892" height="664" alt="Screenshot 2025-10-28 at 10 53 30 PM" src="https://github.com/user-attachments/assets/4130db0e-9d6e-4e10-ae21-98e2b86567cd" />

## Folder Structure Overview

```
Folder/File                     Layer           Responsibility

cmd/main.go                     Entry Point     Application initialization, MongoDB connection, dependency injection, and server start.

internal/domain/url.go          Domain          Defines core entities (domain.URL), requests, and interface contracts (domain.Repository, domain.ShortenerService).

internal/repository/mongo.go    Repository      Implements the domain.Repository contract using the official MongoDB Go driver.

internal/service/shortener.go   Service         Contains business logic (URL validation, unique short code generation, CRUD operations).

internal/handler/http.go        Handler         Handles HTTP routing (using Gorilla Mux), request parsing, error mapping, and response serialization.
```

## Prerequisites

1. **Go**: Version 1.22 or later.

2. **MongoDB**: A running MongoDB instance (default connection: mongodb://localhost:27017).

## Getting Started

**1. Run MongoDB**

Ensure your MongoDB server is running on the default port, or update the MongoURI constant in cmd/main.go.

**2. Build and Run the API**

Initialize Go modules and download dependencies:

```
go mod tidy
```

Run the application:

```
go run cmd/main.go
```


The API will start on http://localhost:8081.

## API Endpoints

All API endpoints are prefixed with /shorten. The redirection endpoint is prefixed with /s.
```
Operation       Method      Path                    Request Body            Success Response

Create          POST        /shorten                {"url": "long_url"}     201 Created + URL object

Retrieve        GET         /shorten/{code}         -                       200 OK + URL object

Update          PUT         /shorten/{code}         {"url": "new_long_url"} 200 OK + Updated URL object

Delete          DELETE      /shorten/{code}         -                       204 No Content

Stats           GET         /shorten/{code}/stats   -                       200 OK + URL object with accessCount

Redirect        GET         /s/{code}               -                       307 Temporary Redirect to original URL
```

## Example Test Flow

### 1. Create a Short URL:

```
curl -X POST http://localhost:8081/shorten -H 'Content-Type: application/json' -d '{"url": "https://www.google.com/search?q=go+mongodb"}'
```

### 2. Retrieve Statistics: (Use the shortCode returned in step 1, e.g., abc123)

```
curl -X GET http://localhost:8081/shorten/abc123/stats
```

### 3. Test Redirection (Increments count):

```
curl -I http://localhost:8081/s/abc123
```
## Acknowledgement
https://roadmap.sh/projects/url-shortening-service
