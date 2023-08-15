all:
	go build -o dist/libnss_http.so -buildmode=c-shared -ldflags="-extldflags '-Wl,-soname,libnss_http.so.2' -s -w"
	go build -o dist/nss_http

install: all
	cp dist/libnss_http.so /lib/libnss_http.so.2
	cp dist/nss_http /sbin/nss_http
	ln -s /sbin/nss_http /sbin/nss_http_sshkey

build-test-container:
	docker build -t nss_http_test:latest -f Dockerfile.test .

test: build-test-container
	go test -v ./...

