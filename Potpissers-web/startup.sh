#!/bin/bash

while true; do
    git fetch origin
    git reset --hard origin/main
    Potpissers-web/./Potpissers-web
    echo Server restarting...
    echo Press CTRL + C to stop.
done
