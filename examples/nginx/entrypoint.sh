#!/bin/sh
sh -c "sleep 30 && pkill nginx" &
nginx
sleep 360