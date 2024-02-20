FROM golang:alpine as genesis-base-builder

WORKDIR /app

RUN apk --no-cache add curl unzip

COPY . .

RUN go mod download
