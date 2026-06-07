# g3-world

A complete Olympia G3 flat-file store snapshot, used as the realistic
integration fixture for the JSON ⇄ flat-file converters.

## Source

- Extracted from `olympia-32/tests/g3/olympia/fixtures/lib.tgz`
  (original snapshot dated 1996-08-31, authored by Rich Skrenta).
- `lib/` is the original G3 store, reproduced byte-for-byte. Treat it as
  **read-only** reference data.

## World facts

- **Turn:** 0 (freshly seeded world; see `lib/times_0`).
- **RNG seed:** `slartibartfast?` (see `lib/randseed`).
- **System players** (`lib/system`): indep=100, gm=200, skill=202.
- **Factions** present under `lib/fact/`: 100, 200, 202, 203, 206, 207.
- Approx. size: ~2 MB of text, dominated by `lib/loc` (~1.1 MB).

## Expected JSON

`json/` (to be added) holds the expected olyg6 JSON store for this same world.
Tests should:

- import `lib/` -> JSON and compare against `json/`,
- export `json/` -> flat and compare against `lib/`,
- verify JSON -> flat -> JSON is lossless.

Because of its size, prefer this case for integration tests; use a small
hand-built world for fast unit tests.
