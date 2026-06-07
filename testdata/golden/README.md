# Golden test data

This tree holds **golden files** used by olyg6 tests, especially the flat-file
⇄ JSON store converters and (later) golden-turn regression tests.

The Go toolchain ignores any directory named `testdata`, so nothing here is
compiled. Tests reference these files via relative paths from the package under
test.

## Layout

Each case lives under `cases/<name>/` and is one of two shapes.

### Converter cases

For testing the flat-file ⇄ JSON store converters. The case mirrors the
original Olympia G3 on-disk layout so we can diff byte-for-byte against the C
engine, and provides the expected JSON store for the same world:

```
cases/
  <name>/
    PROVENANCE.md      # where this fixture came from, seeds, and notes
    lib/               # original G3 flat-file store (read-only reference)
      master           #   id -> kind/file/name index
      loc item skill   #   per-kind entity files
      gate road ship
      unform misc
      fact/<player-id> #   a player box plus its unit (char) boxes
      system players randseed times_0 lore ...
    json/              # expected olyg6 JSON store (same world as lib/)
```

Provide both `lib/` and `json/` so converters can be tested in both directions
without deriving expectations from the code under test:

- **import:** `lib/` -> JSON, compare against `json/`
- **export:** `json/` -> flat-file, compare against `lib/`
- **round-trip:** JSON -> flat -> JSON must be lossless

Example: `g3-world`.

### Generator cases

For locking the output of a generator or algorithm against a known-good result.
Inputs live in `fixtures/`, expected outputs in `golden/`:

```
cases/
  <name>/
    PROVENANCE.md      # source/seeds and how to regenerate the goldens
    fixtures/          # inputs fed to the generator
    golden/            # expected outputs (may hold both flat and JSON goldens)
    <generator>        # optional: committed generator for reproducibility
```

`golden/` may carry both the flat-file and JSON forms of the same world (e.g.
`mapgen` emits `loc`/`gate`/`road` and `loc.json`/`gate.json`/`road.json`). When
a golden was produced by the reference C, commit the generator alongside the
case so it can be regenerated (e.g. `lists/harness.c`).

Examples: `mapgen`, `lists`.

## Conventions

- Keep fixtures **read-only** and as **small** as practical. Prefer a tiny
  hand-made world for unit tests; include one or two real snapshots only when
  realism matters.
- Record provenance in each case's `PROVENANCE.md`: the source snapshot (e.g.
  `olympia-32/run/g3/olympia/lib/`), the game turn, any RNG seeds, and the steps
  to regenerate the goldens — turn outcomes depend on seeds.
- Do not edit golden files to make a test pass. If behavior legitimately
  changes, regenerate the goldens deliberately and review the diff.
