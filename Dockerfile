FROM golang:alpine AS builder

RUN apk update && apk add --no-cache \
  ca-certificates \
  musl-dev \
  gcc \
  build-base \
  git
  
# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Run go mod tidy to ensure go.mod and go.sum are up to date
RUN go mod tidy


# Disable CGO and build the application with the no_cgo tag
RUN CGO_ENABLED=0 go build -tags=no_cgo -o main .


# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/main .

# Build a small image
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dist/main /

COPY ./serviceAccountKey.json ./serviceAccountKey.json
COPY ./flow.json ./flow.json
COPY ./.env ./.env
# Command to run
ENTRYPOINT ["/main"]