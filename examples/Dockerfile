# Stage 1: Building the Go app
FROM golang:1.21.4 as builder

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /app

# Copy the go.mod and go.sum to download all dependencies
COPY go.mod .
COPY go.sum .

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app as a static binary
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o main .

# Stage 2: Copy the Go static executable to Distroless container
FROM gcr.io/distroless/static-debian12

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["./main"]

# Port on which the container should listen
EXPOSE 3000
