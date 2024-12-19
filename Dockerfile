# Stage 1: Build the Go application
FROM golang:1.23.4 AS prod-stage 

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the rest of the application
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./main.go

# Expose the application port
EXPOSE 8080

# Command to run the application binary
ENTRYPOINT ["/api"]



# Stage 2: Development environment with air
FROM golang:1.23.4 AS dev-stage

WORKDIR /app

# Install dependencies
RUN apt-get update && apt-get install -y curl git && apt-get clean

# Install air
RUN go install github.com/air-verse/air@latest

# Install dependencies
COPY go.mod go.sum ./
RUN go mod tidy


# Copy the application files
COPY . .

# Expose the application port
EXPOSE 8080

# Command to run air for development
CMD ["air", "-c", ".air.toml"]
