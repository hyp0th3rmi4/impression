BINARY_NAME=impression
SOURCES=$(wildcard *.go cmd/*.go runtime/*.go)

IMAGE_NAME?=hyp0th3rmi4/impression
IMAGE_TAG?=latest
IMAGE_TAG_TEST=$(IMAGE_TAG)-test



# BUILD TARGETS
	
.PHONY: build
build: bin/$(BINARY_NAME) bin/$(BINARY_NAME).test

# buils the service executable
bin/$(BINARY_NAME): $(SOURCES)
	go build -o bin/$(BINARY_NAME) .

# builds the service with test instrumentation
bin/$(BINARY_NAME).test: $(SOURCES)
	go test -c ./ -cover -covermode=set -coverpkg=./... -o bin/$(BINARY_NAME).test -tags=integration

# DOCKER TARGETS
docker-image: $(SOURCES) Dockerfile
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) --target release .
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG_TEST) --target test .


# TEST TARGETS

# virtual target for unit tests
.PHONY: test
test: coverage.profile

# executes unit tests and computes the coverage 
coverage.profile: coverage.out
	go tool cover -func=coverage.out -o coverage.profile

# runs the unit tests andd capture the cover profile
coverage.out:
	go test ./... -coverprofile=coverage.out




# MISCELLANEOUS TARGETS

.PHONY: clean
clean:
	rm -rf bin

# shows the list of Makefile targets
.PHONY: help
.SILENT: help
help:
	sed 's/$$(BINARY_NAME)/$(BINARY_NAME)/g' Makefile | grep '^[a-zA-Z0-9][a-zA-Z0-9/.\-]*:' | cut -d ':' -f 1


