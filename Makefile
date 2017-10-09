all: clean main

main:
	(mkdir build/)
	go build -i -o build/gowebbench src/main.go

gotest:
	(cd src/ && go test -v)

test: clean main
	./build/gowebbench -t 2 http://www.baidu.com
	./build/gowebbench -t 2 https://www.baidu.com
	./build/gowebbench -t 2 -2 https://http2.golang.org/reqinfo

install:
	cp build/gowebbench /usr/local/bin/gowebbench

uninstall:
	rm /usr/local/bin/gowebbench

clean:
	rm -rf build/


