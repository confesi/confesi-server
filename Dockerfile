# syntax=docker/dockerfile:1

FROM golang:1.22-alpine as build-go
WORKDIR /confesi
COPY . .
EXPOSE 8080
RUN go build -o app main.go

FROM alpine:3.18.2 as production
WORKDIR /confesi
COPY --from=build-go /confesi .
CMD ["./app"]
