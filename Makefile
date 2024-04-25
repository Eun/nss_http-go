
ifdef FORCE_DEBUG_LOG
ADDITIONAL_LD_FLAGS = -X main.ForceDebugLog=true -X main.ForceTraceLog=true
endif

all:
	go build -o dist/libnss_http.so -buildmode=c-shared -ldflags="-extldflags '-Wl,-soname,libnss_http.so.2' -s -w ${ADDITIONAL_LD_FLAGS}"
	go build -o dist/nss_http

install: all
	cp dist/libnss_http.so /lib/libnss_http.so.2
	cp dist/nss_http /sbin/nss_http
	ln -s /sbin/nss_http /sbin/nss_http_sshkey

test-container:
	docker build -t nss_http_test:latest -f Dockerfile.test .

run-test-container: test-container
	docker run --rm -ti nss_http_test:latest

test: test-container
	go test -v

