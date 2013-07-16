all: build

install-deps:
	go get code.google.com/p/go.net/websocket
	go get github.com/tarm/goserial

build: install-deps
	rm -rf build
	mkdir build
	cd src/ && go build -o ../build/puppet-server

install: build
	cp build/puppet-server /usr/local/bin/puppet-server
	chmod 0755 /usr/local/bin/puppet-server

uninstall:
	rm /usr/local/bin/puppet-server
	rm -rf /var/lib/puppet-server

clean:
	rm -rf build/