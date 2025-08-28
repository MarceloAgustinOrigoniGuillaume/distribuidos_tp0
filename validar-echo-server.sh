#!/bin/bash

# -w 2 == 2 seconds of inactivity for timeout.
res=$(echo "probe_message" | netcat 127.0.0.1 1234 -w 2)

if [[ "probe_message" == "$res" ]]; then
	echo "action: test_echo_server | result: success"
else
	echo "action: test_echo_server | result: fail"
fi