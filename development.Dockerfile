# syntax=docker/dockerfile:1

FROM golang:1.20-alpine
WORKDIR /confesi
RUN go install github.com/cosmtrek/air@v1.49
EXPOSE 8080
CMD ["air"]
