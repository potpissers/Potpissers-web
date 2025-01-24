#!/bin/bash

while true; do
    git fetch origin
    git reset --hard origin/main
    ./potpissers-web
    echo Server restarting...
    echo Press CTRL + C to stop.
done