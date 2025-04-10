#!/bin/bash

while true; do
    git fetch upstream
    git reset --hard upstream/main
    /Potpissers-web/./Potpissers-web
    echo Server restarting...
    echo Press CTRL + C to stop.
done
