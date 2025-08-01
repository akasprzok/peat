# syntax=docker/dockerfile:1.17.1

#######################
##### BUILD STAGE #####
#######################
FROM golang:1.24.5 AS builder

WORKDIR /app

# Copy go.mod and go.sum to leverage Docker layer caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 disables CGO, creating a statically linked binary for better portability
# GOOS=linux specifies the target operating system
# -o server specifies the output binary name
RUN CGO_ENABLED=0 GOOS=linux go build -o peat .

#########################
##### RUNTIME STAGE #####
#########################
FROM gcr.io/distroless/static:nonroot AS release

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/peat .

# Expose the port your application listens on (if applicable)
EXPOSE 8080

# Command to run the application when the container starts
CMD ["./peat"]