#!/usr/bin/env bash

# set -x

sed -i -e "s|http://localhost:3000|${SERVER_URL}|g" /app/out/updatePlugins.xml

exec "$@"
