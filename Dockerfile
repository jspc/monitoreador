FROM alpine

ENV GOPATH /go

RUN apk update && \
    apk add git ca-certificates go build-base && \
    update-ca-certificates && \
    go get gopkg.in/gcfg.v1 && \
    go get github.com/guillermo/go.procmeminfo && \
    go get github.com/zenazn/goji && \
    go get github.com/zenazn/goji/web

COPY . /app
WORKDIR /app

RUN adduser -S -h /etc/monitoreador -D monitoreador
RUN make & make install

EXPOSE 8000
CMD monitoreador
