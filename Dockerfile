FROM golang

ENV SRC_DIR=/go/src/github.com/blinky-z/Blog
ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN go build -o server
ENTRYPOINT ["./server"]

EXPOSE 8080