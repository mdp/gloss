EXECUTABLE := gloss

all: clean install build

build:
	go build .

release:
	goxc

install:
	go get github.com/laher/goxc
	go install

clean:
	rm -rf debian releases

.PHONY: clean release dep install
