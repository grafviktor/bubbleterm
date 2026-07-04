#!/bin/bash

set -e

cd "$(dirname "$(readlink -f "$0")")"

vhs "demo.tape" > /dev/null
