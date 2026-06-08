# island

Generator case for `pkg/islandgen`: a blank ocean map + seed fed to the island
generator, with the resulting map and seed locked as goldens.

## Fixtures (`fixtures/`)

- `ascii-map.txt` — a 40×40 all-ocean map (`.`).
- `randseed` — 16-byte RNG seed, bytes `0x01..0x10`. Island shape and terrain
  depend on this seed.

## Goldens (`golden/`)

Output of `oly-g6 generate island` on the fixtures with `--size 60 --border 2
--shelf 3` (the values in `island_test.go`'s `goldenOpts`):

- `ascii-map.txt` — the map with one ~60-province island added.
- `randseed`, `randseed.json` — the advanced RNG state (binary + hex).

Produced by our Go (there is no standalone C island binary in this tree to diff
against); generated once and eyeballed for a plausible, contiguous,
terrain-clustered island away from the border, then locked by `TestGolden`.

## Regenerating

```
mkdir -p /tmp/islgen && cp fixtures/* /tmp/islgen/
go run ./cmd/oly-g6 generate island -i /tmp/islgen -o /tmp/islgen --size 60 --border 2 --shelf 3
cp /tmp/islgen/{ascii-map.txt,randseed,randseed.json} golden/
```

Only regenerate when behavior legitimately changes; review the diff.
