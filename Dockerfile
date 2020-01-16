# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# We specify the base image we need for our go application
# https://hub.docker.com/_/golang?tab=tags
FROM golang:1.13.5 as builder

# Add Maintainer Info
LABEL maintainer="Logogo <lointo@163.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# We copy everything in the root directory into our /app directory
# ADD . /app

# Copy go mod and sum files
COPY go.mod go.sum ./

#set go env  goproxy.io
ENV GOPROXY "https://goproxy.io"

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o member_es_basic_api main.go

# Multi-stage build - final stage
FROM scratch
COPY --from=builder /app/member_es_basic_api /app/
COPY --from=builder /app/conf /app/conf
# Expose port 8080 to the outside world
EXPOSE 8049

# Command to run the executable
ENTRYPOINT ["/app/member_es_basic_api"]
