# gowebbench - refine webbench in C

Travis CI: [![Build Status](https://travis-ci.org/qinyuhang/go-webbench.svg?branch=master)](https://travis-ci.org/qinyuhang/go-webbench)

See [Webbench written in C](https://github.com/qinyuhang/WebBench/)

This is a free software, see GNU Public License version 3 for details.

## Usage
```
./main [-c clientNumber -t requestTimeInSecond --(httpMethod)] url
```

## Show Help
```
./main -h
```

## How To build
```
git clone https://github.com/qinyuhang/go-webbench.git
cd go-webbench
make
```

## TODO
- [ ] Makefile install target

- [x] Force HTTP2 support

- [x] Switch UA support 

- [x] -f switch support Pragma no-cache

- [x] -r switch support --force-reload

- [ ] -F switch support --Field

- [ ] --file
using with POST send file

- [ ] -i switch support --input
read json or csv config

- [ ] -d switch support --data

- [ ] -p switch support --proxy

- [ ] -o json output verbose request info
