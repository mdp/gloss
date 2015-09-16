EXECUTABLE := gloss

all: clean install bin arm6 arm7 linux386 linux darwin

bin:
	go build -o "bin/$(EXECUTABLE)"

# arm
arm6:
	GOARM=6 GOARCH=arm GOOS=linux go build -o "bin/linux/arm/6/$(EXECUTABLE)"
arm7:
	GOARM=7 GOARCH=arm GOOS=linux go build -o "bin/linux/arm/7/$(EXECUTABLE)"

# 386
linux386:
	GOARCH=386 GOOS=linux go build -o "bin/linux/386/$(EXECUTABLE)"

# amd64
darwin:
	GOARCH=amd64 GOOS=darwin go build -o "bin/darwin/amd64/$(EXECUTABLE)"
linux:
	GOARCH=amd64 GOOS=linux go build -o "bin/linux/amd64/$(EXECUTABLE)"

install:
	go install

clean:
	rm -rf bin/

.PHONY: clean release dep install
