# AGENTS.md

Guidance for AI agents and contributors working on **olyg6** ("G6").

## What this project is

olyg6 is a faithful port of the **Olympia G3** play-by-mail (PBM) strategy game
engine from C to Go.

- **Goal:** reproduce G3 *gameplay* behavior, not its C code structure.
- **Faithfulness:** match observable game behavior (turn results, reports,
  random outcomes given the same seeds, data the engine reads/writes).
- **Features:** avoid adding new features. We *may* fix bugs we find in the
  original; when we do, document the bug and the fix.
- **Tests:** writing Go tests is a first-class requirement. Prefer adding tests
  alongside any ported subsystem. Long-term maintainability depends on them.

The CLI is named **`oly-g6`**.

## Reference C source (read-only)

The original C source is **not** part of this repository. It lives at:

```
/Users/wraith/Software/mdhender/olympia-32/g3/olympia/
```

Treat it as a read-only reference. Key files when porting:

- `oly.h` — central entity definitions and `struct box` layout
- `io.c` — flat-file load/save (the store format; see below)
- `code.c` — entity-id <-> display-code encoding (e.g. `10000` <-> `aa00`)
- `main.c` — game loop / turn processing entry point
- `input.c`, `order.c` — order parsing and command dispatch
- `combat.c`, `move.c`, `loc.c` — combat, movement, locations
- `z.c` / `z.h` — utility helpers (ilist, allocation)

When porting a behavior, **read the relevant C first** and ground the Go
implementation in what the C actually does.

## License and attribution

- This Go port is licensed under the **MIT License** (see `LICENSE`),
  Copyright (c) 2026 Michael D Henderson.
- The original Olympia C sources (G1/G2/G3) were released into the
  **public domain** by their author, **Rich Skrenta**. Acknowledge this in
  user-facing docs. G3 upstream: https://github.com/olympiag3/olympiag3

## Tech stack and conventions

- **Module:** `github.com/mdhender/olyg6`
- **Go:** 1.26.x (see `go.mod`)
- **CLI / flags:** use `github.com/peterbourgon/ff/v4` — **not** Cobra.
- **Versioning:** use `github.com/maloquacious/semver`.
- Standard Go layout: commands under `cmd/` (e.g. `cmd/oly-g6/`), library
  packages under the module root or an `internal/` tree as appropriate.
- Run `gofmt` (or `go fmt ./...`) on all Go files. Keep code idiomatic Go;
  do not transliterate C idioms (no global mutable `bx[]` arrays, no
  pointer-as-int tricks). Use explicit Go types for entity ids.

### Build / test / verify

```
go build ./...
go test ./...
go vet ./...
```

Always build and run tests before declaring a change complete.

## The data store: flat-file vs JSON

A deliberate divergence from G3: G6 uses **JSON** as its native store because
it is easier to work with in Go. We also provide **converters** so we can
round-trip to the original flat-file format and diff against the C engine.

```
                    convert (export)
  JSON store  ───────────────────────────▶  G3 flat-file lib/
   (native)   ◀───────────────────────────  (compare vs original engine)
                    convert (import)
```

### Original flat-file format (what the converters target)

The G3 store is a directory (the `lib/` dir) of line-oriented text files plus a
`master` index. Entities are "boxes" identified by an integer id.

- One blank line separates boxes within a file.
- A box starts with: `<id> <kind> <subkind>` (subkind is `0` if none),
  e.g. `10000 loc ocean`.
- Subsequent lines are `KK value`, where `KK` is a two-character key followed by
  one whitespace char, e.g. `na Name`, `il\t<item list>`.
- Nested sections use uppercase two-letter tags (`CH`, `LO`, `LI`, `SL`, `IT`,
  `PL`, `SK`, `GA`, `MI`, `IM`, `CO`, `CM`) introducing a sub-record whose
  following indented `kk value` lines belong to it.
- Box-reference lists are space-separated ids; a trailing `\` means the list
  continues on the next (tab-indented) line.

Files are split by entity kind: `loc`, `item`, `skill`, `gate`, `road`, `ship`,
`unform`, plus `misc` (leftovers), `fact/<player-id>` (a player and its units),
and `master` (the id -> kind/file/name index). System files include `system`,
`players`, `randseed`/`randseed`, `times_0`, `lore`, `gate`, etc.

Entity ids encode a display code (see `code.c`): ids `< 10000` print as
decimal; higher ranges encode as letter+digit codes using the restricted
alphabet `abcdfghjkmnpqrstvwxz` (no vowels except `a`, no `l`).

Entity kinds (`oly.h`): `T_player=1, T_char=2, T_loc=3, T_item=4, T_skill=5,
T_gate=6, T_road=7, T_deadchar=8, T_ship=9, T_post=10, T_storm=11,
T_unform=12, T_lore=13`.

### Converter expectations

- Export to flat-file must be **byte-comparable** (or as close as practical)
  with G3 output so we can diff turn results. Preserve key ordering, spacing,
  list wrapping, and the `master` layout.
- Round-trip (JSON -> flat -> JSON) must be lossless.
- Add tests using real sample data from the reference run dirs, e.g.
  `/Users/wraith/Software/mdhender/olympia-32/run/g3/olympia/lib/`.

## Golden test data

Golden files live under `testdata/golden/` (the Go toolchain ignores any
`testdata` directory, so nothing there is compiled). See
`testdata/golden/README.md` for the full convention. In short, each case lives
in `testdata/golden/cases/<name>/` and is one of two shapes:

- **Converter cases** mirror the original G3 on-disk layout, holding both the
  original flat-file `lib/` and the expected olyg6 `json/` store for the same
  world (e.g. `g3-world`). Used to test the flat ⇄ JSON converters in both
  directions plus round-trip: `lib/` -> JSON (compare `json/`), `json/` -> flat
  (compare `lib/`), and JSON -> flat -> JSON (lossless).
- **Generator cases** hold inputs in `fixtures/` and expected outputs in
  `golden/` — and `golden/` may carry both flat and JSON goldens for the same
  world (e.g. `mapgen`, `lists`). Used to lock a generator/algorithm's output
  against a known-good result. When that result was produced by the reference C,
  keep the generator in the case dir (e.g. `lists/harness.c`).

Rules common to both:

- Keep fixtures small and read-only; record provenance — source snapshot, game
  turn, RNG seeds, and regeneration steps — in each case's `PROVENANCE.md`.
  Outcomes depend on seeds.
- Never edit a golden file just to make a test pass. Regenerate goldens
  deliberately and review the diff when behavior legitimately changes.

## Working agreements

- Read the C reference before porting; cite the source file/function in commit
  messages or comments when behavior is non-obvious.
- Keep changes small and focused. Prefer the smallest correct change.
- When you discover and fix an original bug, note it (a `BUGS.md` entry or a
  code comment referencing the C location) and add a regression test.
- Do not commit or push unless explicitly asked.
- **Branching (alpha):** while the version carries an `-alpha` pre-release
  marker (see `version.go`), commit and push directly to `main` — no feature
  branches or PRs. Once we drop the `-alpha` marker, switch to branch-per-change
  with PRs.
