#!/bin/bash

go run ./cmd/oly-g6 generate map --output-path cmd/oly-g6/testdata/output --input-path cmd/oly-g6/testdata/input -S olympia-g6-randseed-v1.json -R olympia-g6-regions-v1.json -L olympia-g6-lands-v1.json -C olympia-g6-cities-v1.txt -M olympia-g6-ascii-map-v1.txt

exit $?
