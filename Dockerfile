# syntax=docker/dockerfile:1

FROM golang:1.23
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./
COPY /models ./models

RUN go build -o /polcompass

CMD ["/polcompass"]