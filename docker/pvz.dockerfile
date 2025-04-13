FROM golang:1.24.0 AS builder

WORKDIR /usr/src/app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go build -o /usr/src/app/pvz ./cmd

FROM alpine:3.21.3

RUN apk add --no-cache curl postgresql-client bash

COPY wait-for-db.sh /usr/local/bin/wait-for-db
RUN chmod +x /usr/local/bin/wait-for-db

COPY --from=builder /usr/src/app/pvz /pvz

EXPOSE 8080

CMD ["/usr/local/bin/wait-for-db"]
