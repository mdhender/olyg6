# oly-g6

Command-line interface for the Olympia G6 engine.

```
oly-g6 generate canvas   create a blank multi-ocean ASCII map
oly-g6 generate map      generate the world store from an ASCII map
oly-g6 generate island   add a randomly-shaped island to an ASCII map
oly-g6 version           print the engine version
```

All `generate` subcommands read inputs from `--input-path` (`-i`, default `.`)
and write outputs to `--output-path` (`-o`, default `.`). Keeping the two
separate avoids overwriting source files when building a new game.

## The map-building pipeline

The world starts as an **ASCII art map** — a grid of glyphs, one per province.
`generate canvas` makes the blank ocean canvas; `generate island` shapes the
land masses; `generate map` then turns the finished map into the game store:

```
generate canvas ──► generate island ──► … ──► generate island ──► [hand-edit] ──► generate map ──► loc/gate/road (+ .json)
```

Run `generate island` repeatedly to grow continents, hand-edit the ASCII map as
desired, then run `generate map` once.

### Ocean glyphs

Ocean tiles use one of eight glyphs, in four "color" pairs — a plain tile and a
**sea-lane** tile per pair (sea lanes allow fast ocean travel):

| color | plain | sea lane |
|-------|-------|----------|
| 1 | `,` | `;` |
| 2 | `.` | `:` |
| 3 | (space) | `~` |
| 4 | `'` | `"` |

Both subcommands share one definition of "ocean" (`pkg/asciimap`), so they never
disagree about which tiles are water.

## generate canvas

Creates the blank starting map: a square, all-ocean grid partitioned into
several distinct oceans. It is the first step of the pipeline.

```
oly-g6 generate canvas [flags]

  -i, --input-path STRING    directory containing the seed file (default: .)
  -o, --output-path STRING   directory to write output files (default: .)
  -M, --ascii-map STRING     ascii map file to create (default: ascii-map.txt)
  -S, --seed STRING          random seed input file; binary or .json (default: randseed)
      --size INT             square map edge length, 9-99 (default: 99)
      --oceans INT           number of oceans, 1-20 and < size/2 (default: 7)
```

Each ocean is grown from a random seed point and assigned one of the four plain
ocean glyphs (`,` `.` space `'`) so that adjacent oceans differ — including
diagonal contact and the east-west wrap, matching how `generate map` groups
ocean regions. Four glyphs always suffice (at most four oceans meet at any
corner). The output therefore contains **no sea lanes**, so `generate island`
will accept it.

It **creates** the map file and **never reads** one: if the map file already
exists in the output path, the command errors and writes nothing. The seed is
read, advanced, and written (`randseed` + `randseed.json`) so runs are
reproducible and chainable.

## generate island

Adds one random island to the ASCII map and writes the updated map plus the
advanced seed to the output path. A faithful port of G3's `island.c`.

```
oly-g6 generate island [flags]

  -i, --input-path STRING    directory containing input files (default: .)
  -o, --output-path STRING   directory to write output files (default: .)
  -M, --ascii-map STRING     ascii map input/output file (default: ascii-map.txt)
  -S, --seed STRING          random seed input file; binary or .json (default: randseed)
  -b, --border INT           edge border kept clear of new land (default: 2)
  -c, --shelf INT            continental shelf kept around existing land (default: 3)
      --size INT             target island size in provinces; 0 = random (default: 0)
```

It seeds the island at the point farthest from existing land, grows it to the
target size, and fills it with clustered terrain (`p`/`f`/`m`/`d`/`s`). Output is
deterministic for a given map + seed, and the seed is advanced and written
(`randseed` + `randseed.json`) so runs can be chained.

### Refuses maps with sea lanes

If the input map contains **any** sea-lane glyph (`;` `:` `~` `"`), island
**terminates with an error and writes nothing.** Sea lanes are a finishing
touch: by the time they are added, the author has settled the land masses. Since
island reshapes land, it refuses to edit a map that is in its final stages.

```
$ oly-g6 generate island -i ./world -o ./world2
oly-g6: generate island: island: map ascii-map.txt contains a sea lane '~' at row 10 col 9; island only edits maps before sea lanes are added
```

## generate map

Reads the ASCII map plus the cities/lands/regions name files and the seed, and
writes the G3 flat-file store (`loc`, `gate`, `road`, `randseed`) and the G6
native JSON store (`loc.json`, `gate.json`, `road.json`, `randseed.json`).

```
oly-g6 generate map [flags]

  -i, --input-path STRING    directory containing input files (default: .)
  -o, --output-path STRING   directory to write output files (default: .)
  -M, --ascii-map STRING     ascii map input file (default: ascii-map.txt)
  -C, --cities STRING        city names input file (default: cities.txt)
  -L, --lands STRING         land-area names input file (default: lands.json)
  -R, --regions STRING       region names input file (default: regions.json)
  -S, --seed STRING          random seed input file; binary or .json (default: randseed)
```
