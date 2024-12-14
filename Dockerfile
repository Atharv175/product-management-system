# Use the official Go image as a base image
FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy all files into the working directory
COPY . .

# Download Go dependencies
RUN go mod tidy

# Build the Go application
RUN go build -o main .

# Expose the port your app runs on
EXPOSE 8080

# Run the application
CMD ["./main"]
