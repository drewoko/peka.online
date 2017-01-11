FROM golang:1.7-alpine

MAINTAINER Deniss Gubanov <deniss@gubanov.ee>

WORKDIR /root

COPY ./src/com/github/drewoko /root

ENV PATH $GOPATH/src/github.com/jteeuwen/go-bindata/go-bindata:$PATH

RUN apk add --no-cache git build-base && \
    go get -t github.com/jteeuwen/go-bindata && \
    cd $GOPATH/src/github.com/jteeuwen/go-bindata/go-bindata && \
    go build && \
    cd /root/pekaonline && \
    go get -d ./... && \
    cd static && \
    cd .. && \
    go-bindata -o core/bindata.go -pkg core static/* && \
    go build && \
    cp pekaonline /bin/pekaonline && \
    cd /root && \
    rm -rf pekaonline && \
    rm -rf /usr/local/go \
    rm -rf /go  && \
    >application.properties && \
    apk del build-base git

ENTRYPOINT ["pekaonline"]