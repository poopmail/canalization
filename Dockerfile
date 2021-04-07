# Choose the golang image as the build base image
FROM golang:1.16-alpine AS build

# Define the directory we should work in
WORKDIR /app

# Download the necessary go modules
COPY go.mod go.sum ./
RUN go mod download

# Build the application
ARG CANAL_VERSION=unset-debug
COPY . .
RUN go build \
        -o server \
        -ldflags "\
            -X github.com/poopmail/canalization/internal/static.ApplicationMode=PROD \
            -X github.com/poopmail/canalization/internal/static.ApplicationVersion=$CANAL_VERSION" \
        ./cmd/server/

# Run the application in an empty alpine environment
FROM alpine:latest
WORKDIR /root
COPY --from=build /app/server .
CMD ["./server"]