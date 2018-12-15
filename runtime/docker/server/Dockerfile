FROM golang:1.11

COPY ./CompileDaemon /usr/bin/

WORKDIR /app/

CMD CompileDaemon \
    --build='go build -o server' \
    --command='./server' \
    --log-prefix="false"
