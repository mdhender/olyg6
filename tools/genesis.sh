#!/bin/bash

# Remove the old canvas map first. This rm is REQUIRED for re-runs:
# `generate canvas` refuses to overwrite an existing map file and will error if
# ascii-map-canvas.txt is already there.
rm -f cmd/oly-g6/testdata/output/ascii-map-canvas.txt

# create a new map with 9 oceans
#
# Besides the map, `generate canvas` also writes the advanced seed to the output
# dir as both randseed and randseed.json. The island steps below read that
# randseed.json (-S) from the same output dir, so this canvas step MUST run
# first to create it. (If you ever add cleanup for randseed*.json or
# ascii-map-island.txt, keep that ordering in mind.)
go run ./cmd/oly-g6 generate canvas --output-path cmd/oly-g6/testdata/output --input-path cmd/oly-g6/testdata/input -S randseed-canvas.json -M ascii-map-canvas.txt --oceans 9

# Seed the island map from the canvas. ascii-map-island.txt is overwritten in
# place each run by the island steps (island has no overwrite guard, unlike
# canvas), so it does not need a separate rm.
cp -p cmd/oly-g6/testdata/output/ascii-map-canvas.txt cmd/oly-g6/testdata/output/ascii-map-island.txt

# Generate islands, largest first. Each island run reads the map AND seed from
# the output dir and writes them back there (input-path == output-path), so the
# runs chain: every island lands on the accumulated map and advances the shared
# seed. The inner loop adds (i+1) islands of each successive size.
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

