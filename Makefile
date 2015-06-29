# Here we define some variables to be used within
# the Makefile.

GIANTSWARM_USERNAME := $(shell swarm user)
GO_SOURCE := $(shell find . -name '*.go')
GO_PATH := $(shell pwd)/.gobuild
GO_PROJECT_PATH := $(GO_PATH)/src/github.com/giantswarm

# Phony targets are all targets which have names that
# do not resemble some file or folder being generated
# as a result
.PHONY=all clean deps currentweather swarm-up docker-build docker-push

# The default target when you issue 'make'
all: deps currentweather

deps: .gobuild
.gobuild:
	mkdir -p $(GO_PROJECT_PATH)
	cd $(GO_PROJECT_PATH) && ln -s ../../../.. currentweather

	# Fetch public packages
	GOPATH=$(GO_PATH) go get -d github.com/giantswarm/currentweather

# Compiling the Golang binary for Linux from main.go and libraries.
# We actually use another Docker container for this to ensure
# this works even on non-Linux systems.
currentweather: $(GO_SOURCE) 
	echo Building for linux/amd64
	docker run \
	    --rm \
	    -it \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=linux \
	    -e GOARCH=amd64 \
	    -w /usr/code \
	    golang:1.4 \
	    go build -a -o currentweather

# Building your custom docker image
docker-build: currentweather
	docker build -t registry.giantswarm.io/$(GIANTSWARM_USERNAME)/currentweather-go .

# Starting redis container to run in the background
docker-run-redis:
	docker run --name=redis -d redis

# Testing your custom-built docker image locally
docker-run:
	docker run --link redis:redis -p 8080:8080 \
		-ti --rm --name currentweather-go-container \
		registry.giantswarm.io/$(GIANTSWARM_USERNAME)/currentweather-go

# Pushing the freshly built image to the registry
docker-push:
	docker push registry.giantswarm.io/$(GIANTSWARM_USERNAME)/currentweather-go

# Starting your application on Giant Swarm.
# Requires prior pushing to the registry ('make docker-push')
swarm-up:
	swarm up swarm.json --var=username=$(GIANTSWARM_USERNAME)

# Removing your application again from Giant Swarm
# to free resources. Also required before changing
# the swarm.json file and re-issueing 'swarm up'
swarm-delete:
	swarm delete currentweather

# To remove the stuff we built locally afterwards
clean:
	rm -rf $(GO_PATH) currentweather
	docker rmi -f registry.giantswarm.io/$(GIANTSWARM_USERNAME)/currentweather-go
