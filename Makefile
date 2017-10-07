all: clean main

main:
	(mkdir build/)
	go build -i -o build/main src/main.go

gotest:
	(cd src/ && go test -v)

test: clean main
	./build/main

clean:
	rm -rf build/


