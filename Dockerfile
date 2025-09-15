# Use the official Go image as a "builder"
FROM golang:1.24.6-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project source code
COPY . .

# Build a static binary that will not depend on system libraries
# CGO_ENABLED=0 is the key to success for small images
RUN CGO_ENABLED=0 go build -o /app/ssh-fetcher .

# --- Stage 2: Final image ---
# Use a minimalistic image. Alpine is an excellent choice.
FROM alpine:latest

RUN apk add --no-cache curl

# Set the working directory
WORKDIR /app
RUN mkdir logs

# Copy ONLY the compiled binary from the build stage
COPY --from=builder /app/ssh-fetcher .

# Copy the folder with the default configuration
COPY devices/ ./devices/

# Specify the command to be executed when the container starts
# We use an array to avoid problems with shell argument handling
ENTRYPOINT ["./ssh-fetcher"]
