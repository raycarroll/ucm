FROM golang:1.22
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
COPY /data/* /docker/starlight/shared_directory/config_files_starlight/spectrum/
RUN mkdir /docker/starlight/shared_directory/config_files_starlight/spectrum/processed
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /ucm
EXPOSE 5672
CMD ["/ucm"]