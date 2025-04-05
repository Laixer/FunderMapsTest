# FunderMaps API

This repository contains the backend API service for FunderMaps.

## Features

- RESTful API with Go Fiber
- Authentication with JWT
- PostgreSQL database integration
- Environment-based configuration
- API documentation
- Health checks and monitoring

## Requirements

- Go 1.21+
- PostgreSQL 16+

## Getting Started

### Setup Environment Variables

Copy the example environment file:

```bash
cp .env.example .env
```

Modify the values in `.env` to match your local environment.

### Running Locally

1. Install dependencies:

```bash
go mod download
```

2. Run the application:

```bash
go run ./cmd/server/main.go
```

## API Documentation

API documentation is available at `/docs` when running in development mode.

## Deployment

### Building for Production

Build a production binary:

```bash
go build -o fundermaps-api ./cmd/server/main.go
```

### Running in Production

#### Using Binary Directly

```bash
# Set up environment variables
export ENVIRONMENT=production
export DB_HOST=your-db-host
export DB_USER=your-db-user
export DB_PASSWORD=your-db-password
export DB_NAME=your-db-name
export JWT_SECRET=your-jwt-secret

# Run the binary
./fundermaps-api
```

## Project Structure

```
.
├── app/                    # Application code
│   ├── config/             # Configuration management
│   ├── database/           # Database connections
│   ├── handlers/           # HTTP request handlers
│   │   └── management/     # Admin management handlers
│   ├── middleware/         # HTTP middleware
│   └── models/             # Database models
├── cmd/                    # Application entry points
│   └── server/             # API server
├── public/                 # Public static files
├── static/                 # Static assets
├── storage/                # Uploaded files storage
├── .env.example            # Example environment file
└── README.md               # Project documentation
```

## License

[MIT License](LICENSE)
