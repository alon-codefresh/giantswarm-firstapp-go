#!/bin/sh

# make the /tmp folder writable for redis server to create the socket.
# has to be done here. after the volume is mountet as root
chown redis:redis /tmp

# start with exec for correct signal handling
exec redis-server /usr/local/etc/redis/redis.conf