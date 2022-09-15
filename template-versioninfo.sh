#!/bin/bash

set -euo pipefail

if [ "$#" -ne 5 ]; then
    echo "Illegal number of arguments: exactly 5 are required."
    exit 1
fi

YEAR="$1"
VERSION="$2"
MAJOR="$3"
MINOR="$4"
PATCH="$5"

sed \
    -e "s/{{\s*.Year\s*}}/$YEAR/g" \
    -e "s/{{\s*.Version\s*}}/$VERSION/g" \
    -e "s/{{\s*.Major\s*}}/$MAJOR/g" \
    -e "s/{{\s*.Minor\s*}}/$MINOR/g" \
    -e "s/{{\s*.Patch\s*}}/$PATCH/g" \
    versioninfo.json.j2 > \
    versioninfo.json

