FROM golang:1.18

WORKDIR $GOPATH/src/github.com/denuoweb/janus
COPY . $GOPATH/src/github.com/denuoweb/janus
RUN go get -d ./...

CMD [ "go", "test", "-v", "./..."]
