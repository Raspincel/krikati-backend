# Use the official Go image as the base image
FROM golang:1.24.1-alpine

# Set the working directory inside the container
WORKDIR /app

RUN go install github.com/air-verse/air@latest

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Expose the port the app runs on
EXPOSE 3000

# Command to run the executable
CMD ["air", "-c", ".air.toml"]