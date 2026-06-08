# Unplanned features (vs. the "Map generator 1.0" README)

The reference C tree ships a document, `g3/mapgen/README`, titled **"Olympia Map
generator 1.0"**. It describes an ASCII-map dialect and a region-naming scheme
that **never actually shipped** — the real `g3/mapgen/mapgen.c` (which
`oly-g6 generate map` faithfully ports) reads a different set of glyphs and gets
its names from data files instead.

This note records the gaps so the old README isn't mistaken for how our
generator behaves. "Our generator" below means `oly-g6 generate map`
(`pkg/mapgen`); the glyph handling lives in `readMap` (`pkg/mapgen/generator.go`).

## Missing and conflicting glyphs

"Missing" = documented in the 1.0 README but not handled by our generator (an
unknown glyph makes `readMap` fail with `read_map: unknown terrain`).
"Conflict" = our generator accepts the glyph but assigns it a different meaning.

| Glyph | "1.0" README meaning                                  | `oly-g6 generate map`                                          | Status              |
|-------|-------------------------------------------------------|----------------------------------------------------------------|---------------------|
| `@`   | mountain (alt of `^`)                                 | unsupported → unknown-terrain error                            | Missing             |
| `!`   | swamp (alt of `:`)                                    | unsupported → unknown-terrain error                            | Missing             |
| `-`   | steppe                                                | unsupported → unknown-terrain error (no steppe terrain exists) | Missing             |
| `?`   | hidden province (terrain inferred from a neighbor)    | unsupported → unknown-terrain error                            | Missing             |
| `;`   | desert — *and* "sea lane" (the README lists it twice) | ocean, sea lane (color 1)                                      | Conflict            |
| `:`   | swamp                                                 | ocean, sea lane (color 2)                                      | Conflict            |
| `%`   | desert                                                | land province + a scattered (unnamed) city                     | Conflict            |
| `^`   | mountain                                              | mountain **+ Uldim-pass region-boundary flag**                 | Conflict (extended) |

Glyphs that **do** carry over with the documented meaning: the ocean set `~`,
`.`, `,`, and space; `p`/`P` (plain); `f`/`F` (forest); `#` (impassable
"hole"); and `*` (marks a named city on land).

Beyond 1.0, our generator also reads many glyphs the README never mentions —
e.g. `m`/`M` (mountain), `d`/`D` (desert), `s`/`S` (swamp), the extra ocean
glyphs `'` and `"` (our ocean set is four color pairs, a plain and a sea-lane
glyph each), `o` (random terrain), the special markers `v` `{` `}` `[` `]` `O`
(Uldim pass, Summerbridge, Mt. Olympus), and the digits `0`–`9` (start/standard
cities). Those are extensions, not gaps, so they are out of scope for the table
above.

## Region naming scheme (replaced by the region and land files)

The 1.0 README names places by **embedding two-letter codes in the ASCII map**:
you write a code like `AA` (a land region/continent) or `aa` (an ocean — always
lowercase, and never using `o`, since `o` is plains) somewhere inside the region,
list the codes and names in a `Regions` file, and the generator infers the
terrain of the two code-squares from their neighbors. Oceans are kept separate
by varying the ocean glyph so a name's flood-fill stops at the color border.

Our generator drops the in-map codes entirely. Names are **data-driven by
coordinate**: `oly-g6 generate map` reads a regions file (continents and oceans)
and a lands file (named land areas/clumps), each a JSON array of records giving a
`name` and a `row`/`col` (the lands file also carries a `glyph`). The generator
seeds each named region by flooding from that coordinate. So the ASCII map now
carries **only terrain** — no naming glyphs — and the two-letter-code scheme,
the `Regions` file, the lowercase-ocean/`no-o` rule, and the
infer-terrain-under-the-code behavior simply do not exist here. The ocean-glyph
"color border" idea does survive, but as terrain coloring (`pkg/asciimap`), not
as a naming mechanism.
