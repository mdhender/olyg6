# mapgen

Generator case for `pkg/mapgen`: a fixed world fed to the map generator, with
the expected flat-file and JSON outputs locked as goldens.

## Fixtures (`fixtures/`)

Inputs to the generator, using mapgen's default input names:

- `ascii-map.txt`, `cities.txt`, `lands.json`, `regions.json` — the world
  definition (derived from the G3 `g3/mapgen` sample inputs).
- `randseed` — 16-byte RNG seed/state. Map outcomes depend on this seed.

## Goldens (`golden/`)

Expected outputs for the world above:

- `loc`, `gate`, `road`, `randseed` — the **flat-file** store. Produced by the
  reference C generator (`olympia-32/g3/mapgen`) and reproduced byte-for-byte by
  `pkg/mapgen` (see `TestGoldenParity`).
- `loc.json`, `gate.json`, `road.json`, `randseed.json` — the **G6 native JSON**
  store. Produced by `pkg/mapgen` itself (there is no C JSON to compare against);
  generated once and spot-checked field-by-field against the flat store at
  creation, then locked (see `TestJSONStoreGolden`). `TestJSONStoreConsistency`
  asserts the JSON id sets match the flat store per kind, and `TestRandSeedJSON`
  asserts `randseed.json`'s hex decodes to the flat `randseed` bytes.
  `randseed.json` stores the 16-byte RNG state as a lowercase hex string.

## Regenerating

From the repo root, regenerate into a scratch dir and copy back the goldens:

```
mkdir -p /tmp/mapgen && cp fixtures/* /tmp/mapgen/
go run ./cmd/oly-g6 generate map --input-path /tmp/mapgen --output-path /tmp/mapgen
cp /tmp/mapgen/{loc,gate,road,randseed,loc.json,gate.json,road.json,randseed.json} golden/
```

Only do this when behavior legitimately changes; review the diff.
