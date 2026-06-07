# lists

Generator case for `pkg/lists` `Shuffle` (a port of the G3 `ilist_shuffle`):
the shuffle ordering is locked against the reference C output for a fixed seed.

- `fixtures/seed` — 16-byte RNG seed/state (bytes `0x01..0x10`).
- `golden/shuffle-50` — the ordering the **reference C** `ilist_shuffle`
  produces for the list `[0, 1, ..., 49]` starting from that seed, one int
  per line.
- `harness.c` — the C program used to generate `golden/shuffle-50`.

## Regenerating

Compiled against the read-only reference C tree (see `AGENTS.md`):

```
CSRC=/Users/wraith/Software/mdhender/olympia-32
cc -O0 -I"$CSRC/lib" -I"$CSRC/g3/olympia" \
   harness.c "$CSRC/lib/ilist.c" "$CSRC/g3/olympia/rnd.c" -o harness
./harness > golden/shuffle-50   # run from a dir containing the seed as ./seed.bin
```

The C `rnd()` (g3/olympia/rnd.c) re-hashes the 16-byte digest with MD5 and
masks/rejection-samples `digest[0]`; `pkg/prng.Rnd` reproduces it byte-for-byte.
`ilist_shuffle` is Fisher-Yates drawing `rnd(i, len-1)` per position.
