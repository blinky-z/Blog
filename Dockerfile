FROM golang
ENV SRC_DIR=/go/src/github.com/blinky-z/Blog
WORKDIR $SRC_DIR
ADD . .
RUN go build -o serverRun
EXPOSE 8080
ENTRYPOINT ["./serverRun"]