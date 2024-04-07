FROM golang:alpine as build

WORKDIR /scout-src
COPY scout.go scout.go
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
RUN go build -o /scout-build scout.go

FROM alpine:latest

WORKDIR /scout-app
COPY --from=build /scout-build ./scout
ENTRYPOINT ["./scout"]