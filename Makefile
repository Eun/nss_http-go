all:
	go build -o libnss_http.so -buildmode=c-shared -ldflags="-extldflags '-Wl,-soname,libnss_http.so.2' -s -w"

install: all
	cp libnss_http.so /lib/libnss_http.so.2
