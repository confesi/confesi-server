# syntax=docker/dockerfile:1

FROM golang:1.20-alpine
WORKDIR confesi
COPY . .
EXPOSE 8080
RUN go build -o app app.go
CMD ["./app"]
