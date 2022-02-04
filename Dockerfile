FROM golang:1.14-alpine

RUN apk add --no-cache make gcc musl-dev git

WORKDIR $GOPATH/src/github.com/htmlcoin/janus
COPY ./ $GOPATH/src/github.com/htmlcoin/janus

RUN go build \
        -ldflags \
            "-X 'github.com/htmlcoin/janus/pkg/params.GitSha=`git rev-parse HEAD`'" \
        -o $GOPATH/bin $GOPATH/src/github.com/htmlcoin/janus/... && \
    rm -fr $GOPATH/src/github.com/htmlcoin/janus/.git

ENV HTMLCOIN_RPC=http://htmlcoin:testpasswd@localhost:4889
ENV HTMLCOIN_NETWORK=auto

EXPOSE 24889
EXPOSE 24890

ENTRYPOINT [ "janus" ]
