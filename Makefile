all:
	go build -o dist/libnss_http.so -buildmode=c-shared -ldflags="-extldflags '-Wl,-soname,libnss_http.so.2' -s -w"
	go build -o dist/nss_http

install: all
	cp dist/libnss_http.so /lib/libnss_http.so.2
	cp dist/nss_http /sbin/nss_http
	ln -s /sbin/nss_http /sbin/nss_http_sshkey

test:
	docker compose run test

test-in-docker:
	python3 -m http.server 8000 &
	cp /nss_http/test/nss_http.json /etc/
	cp /nss_http/test/nsswitch.conf /etc/
	sleep 1
	NSS_HTTP_LOG_FILE=/dev/stdout NSS_HTTP_DEBUG=true getent passwd joe
