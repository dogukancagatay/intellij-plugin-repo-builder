#!/usr/bin/env bash

# set -x

sed -i -e "s|http://localhost:3000|${SERVER_URL}|g" /usr/share/nginx/html/updatePlugins.xml

exec "$@"
