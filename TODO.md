# TODO

## Done

- **Map generation MVP.** `oly-g6 generate canvas | island | map` produces the
  world: the G3 flat store (`loc`/`gate`/`road`, golden-tested byte-for-byte)
  and the G6 native JSON store (`loc.json`/`gate.json`/`road.json` + seed).

## Next goal: process a simple order file

First vertical slice of the engine — enough to take a hand-written order file
with a `move` and produce the result. Three pieces:

1. **Add a player (and a unit to move).**
   - In Olympia, a *player* (faction) owns *characters/units*; **units move**,
     not the faction. So we need a player faction plus at least one unit placed
     at a starting location.
   - Define the JSON store entity types for these (extend `pkg/store`): a player
     and a character/unit, with at least: id, kind, name, faction owner, and
     current location (`where`). Keep it minimal — only fields `move` needs.
   - Decide how the player/unit gets created for now: a hand-written JSON seed
     file we load, vs. a small `oly-g6` command to add one. A fixture file is
     probably the fastest path for the slice.

2. **Read the order file.**
   - Olympia order-file shape: a `unit <id>` header line, then that unit's
     order lines (e.g. `move <dir>`), repeated per unit. Confirm the exact
     grammar against the C before parsing.
   - Add an order parser (new package, e.g. `pkg/orders`) that yields
     `{unit, command, args}` records. Start with just enough to parse `move`.

3. **Implement the `move` command.**
   - Move a unit from its province to an adjacent one. Reuse the world the map
     generator already emits: province exits are `loc.json` `prov_dest`
     (N/E/S/W); containment is `where`/`here`.
   - Resolve a direction to the destination province, update the unit's `where`
     (and the old/new locations' `here` lists), and report the move.
   - Scope for the slice: province-to-province moves only. Defer roads, gates,
     sublocation entry/exit, movement cost/time, stacking, and failure cases.

## Grounding (read the C first, per AGENTS.md)

Reference C at `/Users/wraith/Software/mdhender/olympia-32/g3/olympia/`:
- `oly.h` — entity definitions (`T_player`, `T_char`, `struct box`,
  `entity_player`, `entity_char`).
- `input.c`, `order.c` — order parsing and command dispatch.
- `move.c`, `loc.c` — movement and location logic.
- `main.c` — turn-processing entry point (for overall flow; we only need a sliver).

Our side:
- `pkg/store` — JSON entity types (`Location`, `Gate`, `Road`, `RandSeed`); add
  player/unit here, and a loader for the JSON store the engine will read.
- No engine package yet — this slice creates the first one.

## Open questions for the morning

- One combined `oly-g6` subcommand for the turn (e.g. `oly-g6 run` / `oly-g6
  turn`) reading the JSON store + order file, or smaller steps?
- Player/unit: seed via a committed JSON fixture, or add a `generate`/`add`
  command to create one?
- Order grammar: confirm header + `move` syntax and the direction tokens from
  the C before locking the parser.
- Output: how to report the move (turn-report text vs. just the updated store)?
  For the slice, updating the store + a terse log is probably enough.
- Build a golden case for the slice (order file + starting store → expected
  resulting store) once the shape settles.
