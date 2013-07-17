all: build

install-deps:
	go get code.google.com/p/go.net/websocket
	go get github.com/tarm/goserial

build: install-deps
	rm -rf build
	mkdir build
	cd src/ && go build -o ../build/hifi-puppet

# install: build
# 	cp build/hifi-puppet /usr/local/bin/hifi-puppet
# 	chmod 0755 /usr/local/bin/hifi-puppet

# uninstall:
# 	rm /usr/local/bin/hifi-puppet
# 	rm -rf /var/lib/hifi-puppet

clean:
	rm -rf build/