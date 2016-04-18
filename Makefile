CONF=/etc/monitoreador
GOARGS=-a -installsuffix cgo -x
GOENV=CGO_ENABLED=0
USER=monitoreador
BINDIR=/usr/local/sbin
EXECUTABLE=monitoreador

all: main

main:
	$(GOENV) go build $(GOARGS) -o main

config: user
	install -d -m 0755 -o $(USER) $(CONF)
	install -b -m 0644 -o $(USER) config/sample.ini $(CONF)/config.ini

user:
	-useradd --system --home-dir $(CONF) $(USER)

install: all config
	install -d $(BINDIR)
	install -s -m 0750 -o $(USER) main $(BINDIR)/$(EXECUTABLE)

uninstall:
	rm -rfv $(CONF)
	rm -v $(BINDIR)/$(EXECUTABLE)

clean:
	rm -v main
