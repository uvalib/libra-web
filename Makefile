GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOFMT = $(GOCMD) fmt
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOVET = $(GOCMD) vet

BASE_NAME=libra3

build: darwin web

all: darwin linux web

linux-full: linux web

darwin-full: darwin web

web:
	mkdir -p bin/
	cd frontend && npm install && npm run build
	rm -rf bin/public
	mv frontend/dist bin/public

darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/$(BASE_NAME).darwin backend/*.go

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/$(BASE_NAME).linux backend/*.go

clean:
	rm -rf bin

fmt:
	cd backend; $(GOFMT)

vet:
	cd backend; $(GOVET)

dep:
	cd frontend && npm upgrade
	$(GOGET) -u ./backend/...
	$(GOMOD) tidy
	$(GOMOD) verify
