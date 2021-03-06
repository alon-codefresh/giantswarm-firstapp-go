# Getting started with Go, Docker, and Redis on Giant Swarm

This is a simple example that demonstrates how to write a microservice in Golang and deploy it on [Giant Swarm](https://giantswarm.io/). It pings an external API and caches the data in an Redis cache.

Check out the full tutorial here:

https://docs.giantswarm.io/guides/your-first-service/golang/

## Prerequisites

* Have a Giant Swarm account and the [swarm CLI](https://docs.giantswarm.io/reference/cli/) running. [Request a free invite](https://giantswarm.io/).
* Have [Docker](https://docs.docker.com/installation/) installed and running. You should be familiar with the basic Docker commands and how to handle Makefiles.
* _Optional:_ If not in GNU/Linux environment you need [boot2docker](http://boot2docker.io) as well.

## Go Code

The service logic is implemented in [main.go](main.go). It creates a webserver and on root request pings the [openweather API](http://api.openweathermap.org/data/2.5/weather?q=Cologne), caches the result in Redis and extracts and returns the current weather for Cologne.

## Testing the service locally

To run the two required containers locally you just have to do

```
$ make docker-build
$ make docker-run-redis
$ make docker-run
```

This
* builds the Go project into a linux binary
* creates a custom Docker image with the linux binary
* starts both the custom Docker container and a Redis container.

To test it on a Mac run something like: `curl $(boot2docker ip):8080`, on Linux `curl localhost:8080` should be sufficient.

## Running the service on Giant Swarm

To deploy this service on Giant Swarm you just have to do

```
$ make docker-push
$ swarm up
```

This

* uploads the Docker image to the Giant Swarm registry
* creates the service according to the definition in `swarm.json`
* starts the service

To test it run something like: `curl currentweather-YOURUSERNAME.gigantic.io` and replace YOURUSERNAME with your Giant Swarm username.

For all build and deploy details see the [Makefile](Makefile).

For further documentation and guides see the [docs](https://docs.giantswarm.io/).

## In other languages

* [NodeJS](https://github.com/giantswarm/giantswarm-firstapp-nodejs)
* [Ruby](https://github.com/giantswarm/giantswarm-firstapp-ruby)
* [Python](https://github.com/giantswarm/giantswarm-firstapp-python)
* [PHP](https://github.com/giantswarm/giantswarm-firstapp-php)
* [Java](https://github.com/giantswarm/giantswarm-firstapp-java)

## Open Weather API

The weather data is provided by http://openweathermap.org
