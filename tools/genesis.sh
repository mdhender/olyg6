#!/bin/bash

# remove any old maps
rm -f cmd/oly-g6/testdata/output/ascii-map-canvas.txt

# create a new map with 9 oceans
go run ./cmd/oly-g6 generate canvas --output-path cmd/oly-g6/testdata/output --input-path cmd/oly-g6/testdata/input -S randseed-canvas.json -M ascii-map-canvas.txt --oceans 9

cp -p cmd/oly-g6/testdata/output/ascii-map-canvas.txt cmd/oly-g6/testdata/output/ascii-map-island.txt

# generate islands
sizes=(521 257 257 127 61 61 31 17 11 7)
for i in "${!sizes[@]}"; do
    n="${sizes[i]}"

    for ((j=0; j<=i; j++)); do
        go run ./cmd/oly-g6 generate island \
            --output-path cmd/oly-g6/testdata/output \
            --input-path cmd/oly-g6/testdata/output \
            -S randseed.json \
            -M ascii-map-island.txt \
            -b 3 \
            -c 5 \
            --size "$n"
    done
done

exit $?

