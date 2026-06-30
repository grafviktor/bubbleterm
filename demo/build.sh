#!/bin/bash

set -e

cd "$(dirname "$(readlink -f "$0")")"

# function cleanup() {
#     rm -rf hosts.yaml state.yaml app.log .ssh themes 2>/dev/null
# }

# cleanup
vhs "demo.tape" > /dev/null
# cleanup