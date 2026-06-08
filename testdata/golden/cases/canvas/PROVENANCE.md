# canvas

Generator case for `pkg/canvasgen`: a seed fed to the canvas generator, with the
resulting multi-ocean map and advanced seed locked as goldens.

## Fixtures (`fixtures/`)

- `randseed` — 16-byte RNG seed, bytes `0x01..0x10`. Ocean placement, growth,
  and the resulting partition depend on this seed.

## Goldens (`golden/`)

Output of `oly-g6 generate canvas` on the fixture seed with `--size 40
--oceans 5` (the values in `canvas_test.go`'s `goldenOpts`):

- `ascii-map.txt` — a 40×40 all-ocean map partitioned into 5 oceans, 4-colored
  with the plain ocean glyphs (`,` `.` space `'`).
- `randseed`, `randseed.json` — the advanced RNG state (binary + hex).

Produced by our Go (this is a new command, not a port — there is no C reference
to diff against); generated once and verified: `TestRegionCount` confirms the
map yields exactly 5 regions under mapgen's flood rule, and the output is
sea-lane-free.

## Regenerating

```
mkdir -p /tmp/cvgen && cp fixtures/randseed /tmp/cvgen/
go run ./cmd/oly-g6 generate canvas -i /tmp/cvgen -o /tmp/cvgen --size 40 --oceans 5
cp /tmp/cvgen/{ascii-map.txt,randseed,randseed.json} golden/
```

Only regenerate when behavior legitimately changes; review the diff.
