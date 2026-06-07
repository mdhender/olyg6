# olyg6 (G6)

A faithful Go port of the **Olympia G3** play-by-mail (PBM) strategy game
engine.

The goal of olyg6 is to reproduce G3 *gameplay* — the same turn results given
the same inputs — rather than to mirror the original C code structure. We avoid
adding new features, fix bugs as we find them, and treat automated tests as a
core part of the project.

## Status

Early port. Subsystems are being moved over from the original C engine
incrementally, each with Go tests.

## The CLI: `oly-g6`

olyg6 ships as a command-line tool named **`oly-g6`**. Build it with:

```sh
go build ./cmd/oly-g6
```

Command-line parsing uses [`github.com/peterbourgon/ff/v4`](https://github.com/peterbourgon/ff)
and versioning uses [`github.com/maloquacious/semver`](https://github.com/maloquacious/semver).

## Data store: JSON, with flat-file converters

Olympia G3 stores its game database as a directory of line-oriented text files.
olyg6 instead uses **JSON**, which is easier to work with in Go.

To let us validate behavior against the original engine, olyg6 includes tools to
convert between the JSON store and the original flat-file format:

```
   JSON store  ⇄  G3 flat-file lib/  (compare against the original G3 engine)
```

Round-tripping is lossless, and exported flat-files are kept as close to the
original byte layout as practical so turn output can be diffed directly.

## Development

```sh
go build ./...
go test ./...
go vet ./...
```

See [AGENTS.md](AGENTS.md) for porting conventions, the store-format details,
and where the reference C source lives.

## License and acknowledgements

The olyg6 Go port is licensed under the [MIT License](LICENSE),
Copyright (c) 2026 Michael D Henderson.

Olympia was created by **Rich Skrenta**, who released the original G1/G2/G3 C
sources into the **public domain**. This project is a port of that work and
gratefully acknowledges his authorship. The G3 upstream sources are at
<https://github.com/olympiag3/olympiag3>.
