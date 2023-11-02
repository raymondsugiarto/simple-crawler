FROM golang:1.19

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ./main

RUN pwd
RUN ls -l

RUN /app/main https://autify.com

RUN ls -l /app/html
# CMD ["/app/main"]