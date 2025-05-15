all: build

export PATH := /usr/lib/go-1.23/bin/:$(PATH)

build:
	go build ./...