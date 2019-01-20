FROM golang:1.11

COPY ./CompileDaemon /usr/bin/

WORKDIR /app/

COPY ./install_vendors.sh /install_vendors.sh

CMD /install_vendors.sh && CompileDaemon \
    --build='go build -o server' \
    --command='./server' \
    --log-prefix="false"
