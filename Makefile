GOARGS=-a -installsuffix cgo -x
GOENV=CGO_ENABLED=0
BINDIR=/usr/local/sbin
EXECUTABLE=monitoreador
CONTAINER_TAG=quay.io/financialtimes/monitoreador:latest

all: main

main:
	$(GOENV) go build $(GOARGS) -o main

install: all
	install -d $(BINDIR)
	install -s -m 0750 -o $(USER) main $(BINDIR)/$(EXECUTABLE)

build: all
	docker build -t $(CONTAINER_TAG) .

push:
	docker push $(CONTAINER_TAG)

uninstall:
	rm -rfv $(CONF)
	rm -v $(BINDIR)/$(EXECUTABLE)

clean:
	rm -v main
