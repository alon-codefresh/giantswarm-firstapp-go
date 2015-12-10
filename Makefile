# Variables to be used within the Makefile
GIANTSWARM_USERNAME := $(shell swarm user)
GO_SOURCE := $(shell find . -name '*.go')
GO_PATH := $(shell pwd)/.gobuild
GO_PROJECT_PATH := $(GO_PATH)/src/github.com/giantswarm

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
	docker build -t registry02.giantswarm.io/$(GIANTSWARM_USERNAME)/currentweather-go .

# Starting redis container to run in the background
docker-run-redis:
	@docker kill currentweather-redis-container > /dev/null || true
	@docker rm currentweather-redis-container > /dev/null || true
	docker run -d --name currentweather-redis-container redis

# Running your custom-built docker image locally
docker-run:
	docker run --link currentweather-redis-container:redis -p 8080:8080 \
		-ti --rm --name currentweather-go-container \
		registry02.giantswarm.io/$(GIANTSWARM_USERNAME)/currentweather-go

# Pushing the freshly built image to the registry
docker-push:
	docker push registry02.giantswarm.io/$(GIANTSWARM_USERNAME)/currentweather-go

# Starting your service on Giant Swarm.
# Requires prior pushing to the registry ('make docker-push')
swarm-up:
	swarm up

# Removing your service again from Giant Swarm
# to free resources. Also required before changing
# the swarm.json file and re-issueing 'swarm up'
swarm-delete:
	swarm delete currentweather

# To remove the stuff we built locally afterwards
clean:
	rm -rf $(GO_PATH) currentweather
	docker rmi -f registry02.giantswarm.io/$(GIANTSWARM_USERNAME)/currentweather-go
