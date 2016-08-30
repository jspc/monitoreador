FROM alpine

ENV GOPATH /go

RUN apk update && \
    apk add git ca-certificates go build-base && \
    update-ca-certificates && \
    go get github.com/guillermo/go.procmeminfo \

COPY . /app
WORKDIR /app

RUN adduser -S -h /tmp/ -D monitoreador
RUN make & make install

EXPOSE 8000
CMD monitoreador
