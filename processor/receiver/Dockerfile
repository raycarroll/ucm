FROM golang:1.22
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
RUN pwd 
COPY ./grid_example.in /docker/starlight/config_files_starlight/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /receiver
EXPOSE 5672
CMD ["/receiver"]