#!/bin/bash

set -e

cd "$(dirname "$(readlink -f "$0")")"

vhs "termview-multiplexer.tape" > /dev/null
