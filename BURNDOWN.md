# Burndown

Known issues and deferred work. See `AGENTS.md` for how we handle ported bugs
(document the bug + fix, add a regression test).

## Open

### mapgen: seed-sensitive ilist bug (deferred)

A known **seed-sensitive bug in G3 mapgen**, faithfully reproduced in the port.
It surfaces as a nil-pointer panic during `makeRoads` →
`bridgeMountainPorts` → `bridgeMountainSup` (`pkg/mapgen/roads.go:272`) for some
seeds; the canonical `slartibartfast?\n` seed is unaffected.

- **Status:** fix handed upstream (2026-06-11); awaiting their push, then we
  port it. The upstream prompt asks them to change the return in
  `adjacent_tile_water` (C `mapgen/mapgen.c:1121`) to
  `(p && p->terrain == terr_ocean) ? p : NULL`, which is output-neutral for all
  currently-passing seeds (it only converts the slot-8 crash into a valid secret
  sea route), and to regenerate goldens plus add a reproducing-seed regression
  test. The C `adjacent_tile_terr` (`mapgen.c:1138`) has the identical
  off-by-one but is off the crash path and would churn terrain goldens, so it was
  flagged as an optional, separate upstream commit — if upstream skips it, we
  inherit the same latent twin in `adjacentTileTerr` (`generator.go:388`) and
  should keep tracking it here. Until upstream lands, the port still reproduces
  the panic; do not "fix" it locally ahead of them.
- **Root cause (confirmed, faithful to C):** `adjacentTileWater`
  (`generator.go:360`) ends with `if i < MaxDir { return p }; return nil`,
  mirroring the C `(i < MAX_DIR) ? p : NULL`. When the ocean-adjacent direction
  is found on the **last** shuffled slot `dirVector[8]`, the loop sets `p` to the
  ocean tile but `i` reaches `MaxDir`, so it returns `nil` instead of the tile it
  just found (off-by-one). `bridgeMountainSup` then dereferences that `nil`
  (`to.Terrain`). The C `bridge_mountain_sup` does `assert(to->terrain == ...)`,
  so the C segfaults on the same seeds.
- **Why seed-sensitive:** `randomizeDirVector` (a Fisher-Yates shuffle — the
  "ilist shuffle" connection) decides which slot the ocean direction lands in. A
  mountain port whose only ocean neighbor is shuffled into slot 8, gated by
  `rnd(1,7)==7`, only triggers on some seeds.
- **Not the cause:** the row/col wrap in `adjacentTileSup` matches the C exactly
  (col wraps via `MaxColUsed`, row bounded by 99); no axis swap on this path.
- **When fixing:** ground the fix in the C source, document the divergence, and
  add a regression test using a seed that reproduces it (e.g. the byte-pair-
  swapped variant of the canonical seed).
