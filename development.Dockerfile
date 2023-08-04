# syntax=docker/dockerfile:1

FROM golang:1.20-alpine
WORKDIR /confesi
COPY . .
RUN go install github.com/cosmtrek/air@latest
EXPOSE 8080
CMD ["air"]
