FROM golang:1.22
WORKDIR /app

RUN mkdir -p /docker/starlight/config_files_starlight
COPY grid_example.in /docker/starlight/config_files_starlight

COPY go.mod go.sum ./
COPY . . 

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /ucm
EXPOSE 5672
CMD ["/ucm"]