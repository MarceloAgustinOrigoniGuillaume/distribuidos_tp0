#!/bin/bash

NETWORK_NAME="tp0_testing_net"
SERVER_CONTAINER_NAME="server"
SERVER_PORT=1234
MESSAGE="probe_message"

# Run busybox container that has netcat... on the same network as tp0. 
res=$(docker run --rm --network "$NETWORK_NAME" busybox sh -c "echo '$MESSAGE' | nc $SERVER_CONTAINER_NAME $SERVER_PORT -w 1")

if [[ "$res" == "$MESSAGE" ]]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi