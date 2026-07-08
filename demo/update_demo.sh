#!/bin/bash

set -e

cd "$(dirname "$(readlink -f "$0")")"

vhs "terminal-multiplexer.tape" > /dev/null
