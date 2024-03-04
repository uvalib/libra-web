GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOFMT = $(GOCMD) fmt
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOVET = $(GOCMD) vet

BASE_NAME=libra-web

build: darwin web deploy-data

all: darwin linux web deploy-data

linux-full: linux web deploy-data

darwin-full: darwin web deploy-data

deploy-data:
	mkdir -p bin/
	rm -rf bin/data
	mkdir -p bin/data
	cp ./data/* bin/data

web:
	mkdir -p bin/
	cd frontend && npm install && npm run build
	rm -rf bin/public
	mv frontend/dist bin/public

darwin:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -a -o bin/$(BASE_NAME).darwin backend/*.go

linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/$(BASE_NAME).linux backend/*.go

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
