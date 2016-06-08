#!/usr/bin/env bash

killall -9 stock-filter
sleep 1
./stock-filter &

echo "stock-filter started!"
