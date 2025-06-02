MAJOR_VERSION := 0
MINOR_VERSION := 1
VERSION := ${MAJOR_VERSION}.${MINOR_VERSION}

build::
	go build -o gim

test:: build
	./gim regression.txt
