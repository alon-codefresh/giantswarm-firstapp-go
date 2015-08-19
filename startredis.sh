#!/bin/sh
mkdir /tmp2
chown redis:redis /tmp2
#TODO exec
redis-server /usr/local/etc/redis/redis.conf