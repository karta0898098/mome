# build stage
FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod /
COPY go.sum /

RUN go mod download
RUN go mod verify

COPY . .

RUN make linux

RUN ls

# runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk add --no-cache zsh

COPY --from=builder /app/mome /app/main

WORKDIR /app

ENTRYPOINT ["./main"]