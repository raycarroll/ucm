FROM golang:1.21
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /watcher
EXPOSE 5672
CMD ["/watcher"]