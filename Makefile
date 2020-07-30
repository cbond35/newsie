BIN=newsie

export GO111MODULE=on

all: build

clean:
	go clean

build: clean
	go build -o $(BIN)
