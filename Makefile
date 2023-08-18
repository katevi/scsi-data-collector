#!/usr/bin/make -f

help: 
	@echo "make help -- Displays this help"
	@echo "make build -- Build scsi-data-collector"
	@echo "make run -- Compiles and runs scsi-data-collector"
	@echo "make deps -- Install dependencies"
	@echo "make clean -- Cleans build and run artefacts"

deps:
	go mod tidy

build:
	go build -o bin/scsi-data-collector

run: 
	go run main.go

clean:
	rm -rf bin/	data/ scsi-data-collector.log
