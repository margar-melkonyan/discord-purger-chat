FROM golang:1.22 AS builder

WORKDIR /usr/local/src

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . ./
RUN go build -o ./app main.go