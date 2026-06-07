package mapgen

import (
	"bufio"
	"os"
	"strings"
)

// readLines reads the named input file and returns its lines with trailing
// newlines stripped, mirroring how the C code reads its configuration files.
func (g *Generator) readLines(name string) ([]string, error) {
	f, err := os.Open(g.inPath(name))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	r := bufio.NewReader(f)
	for {
		line, rerr := r.ReadString('\n')
		if line == "" && rerr != nil {
			break
		}
		lines = append(lines, strings.TrimRight(line, "\n"))
		if rerr != nil {
			break
		}
	}
	return lines, nil
}

func (g *Generator) ishuffle(l []int) {
	n := len(l) - 1
	for i := 0; i < n; i++ {
		r := g.rnd(i, n)
		if r != i {
			l[i], l[r] = l[r], l[i]
		}
	}
}

func (g *Generator) tshuffle(l []*Tile) {
	n := len(l) - 1
	for i := 0; i < n; i++ {
		r := g.rnd(i, n)
		if r != i {
			l[i], l[r] = l[r], l[i]
		}
	}
}

// ---------------------------------------------------------------------------
// islands
// ---------------------------------------------------------------------------

func (g *Generator) makeIslands() {
	g.NumIslands = g.WaterCount / 3

	for i := 1; i <= g.NumIslands; i++ {
		row := g.rnd(0, g.MaxRowUsed)
		col := g.rnd(0, g.MaxColUsed)

		if g.Map[row][col] != nil && g.Map[row][col].Terrain == TerrOcean &&
			g.islandAllowed(row, col) {
			hidden := g.rnd(0, 1)
			g.createASubloc(row, col, hidden, TerrIsland)
		} else {
			i--
		}
	}
}

func (g *Generator) islandAllowed(row, col int) bool {
	inside := g.Map[row][col].Inside
	if inside == 0 {
		return true
	}
	if strings.Contains(g.InsideNames[inside], "Deep") {
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// graveyards
// ---------------------------------------------------------------------------

func (g *Generator) makeGraveyards() {
	for i := 1; i <= g.InsideTop; i++ {
		p := g.InsideList[i][0]
		if p.Terrain == TerrOcean {
			continue
		}
		n := len(g.InsideList[i])
		if n < 10 {
			continue
		}

		l := make([]*Tile, n)
		copy(l, g.InsideList[i])
		g.tshuffle(l)

		for j := 0; j < n/10; j++ {
			g.createAGraveyard(l[j].Row, l[j].Col)
		}
	}
}

// ---------------------------------------------------------------------------
// sublocation placement
// ---------------------------------------------------------------------------

func (g *Generator) placeSublocations() {
	var l []int

	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			if g.Map[row][col] != nil && g.Map[row][col].Terrain != TerrOcean {
				l = append(l, row*1000+col)
			}
		}
	}

	g.ishuffle(l)

	for i := 0; i < len(l); i++ {
		row := l[i] / 1000
		col := l[i] % 1000

		// Put a city everywhere there is a * or every 1 in 15 locs,
		// randomly. Don't put one where there already is a city (city != 2).
		if g.Map[row][col].City == 1 ||
			(g.rnd(1, 15) == 1 && g.Map[row][col].City != 2) {
			g.createACity(row, col, "", false, 0)
		}

		if g.rnd(1, 100) <= 35 {
			g.makeAppropriateSubloc(row, col)
		}
		if g.rnd(1, 100) <= 35 {
			g.makeAppropriateSubloc(row, col)
		}
		if g.rnd(1, 100) <= 35 {
			g.makeAppropriateSubloc(row, col)
		}
	}
}

// locEntry mirrors the C loc_table[] rows.
type locEntry struct {
	terr   int // terrain appropriate
	kind   int // what to make there
	weight int // weight given to selection
	hidden int // 0=no, 1=yes, 2=rnd(0,1)
}

var locTable = []locEntry{
	{TerrDesert, TerrCave, 10, 1},
	{TerrDesert, TerrOasis, 10, 2},
	{TerrDesert, TerrSandPit, 10, 2},

	{TerrMountain, TerrRuins, 10, 1},
	{TerrMountain, TerrCave, 10, 1},
	{TerrMountain, TerrYewGrove, 8, 2},
	{TerrMountain, TerrTreeCir, 5, 1},
	{TerrMountain, TerrPits, 5, 1},
	{TerrMountain, TerrLair, 10, 2},
	{TerrMountain, TerrBattlefield, 6, 2},

	{TerrSwamp, TerrBog, 10, 2},
	{TerrSwamp, TerrPits, 10, 2},
	{TerrSwamp, TerrBattlefield, 6, 2},
	{TerrSwamp, TerrLair, 5, 2},

	{TerrForest, TerrRuins, 10, 1},
	{TerrForest, TerrRockyHill, 10, 0},
	{TerrForest, TerrTreeCir, 10, 1},
	{TerrForest, TerrEnchFor, 8, 1},
	{TerrForest, TerrPasture, 5, 0},
	{TerrForest, TerrYewGrove, 10, 2},
	{TerrForest, TerrCave, 10, 1},
	{TerrForest, TerrGrove, 9, 1},
	{TerrForest, TerrBattlefield, 6, 2},
	{TerrForest, TerrLair, 3, 1},

	{TerrPlain, TerrRuins, 10, 1},
	{TerrPlain, TerrPasture, 10, 0},
	{TerrPlain, TerrRockyHill, 10, 0},
	{TerrPlain, TerrSacGrove, 6, 2},
	{TerrPlain, TerrTreeCir, 6, 1},
	{TerrPlain, TerrPopField, 10, 0},
	{TerrPlain, TerrCave, 10, 1},
	{TerrPlain, TerrBattlefield, 6, 2},
}

func (g *Generator) makeAppropriateSubloc(row, col int) {
	terr := g.Map[row][col].Terrain

	sum := 0
	for i := range locTable {
		if locTable[i].terr == terr {
			sum += locTable[i].weight
		}
	}

	if sum <= 0 {
		g.logf("no subloc appropriate for (%d,%d)\n", row, col)
		return
	}

	n := g.rnd(1, sum)

	for i := range locTable {
		if locTable[i].terr != terr {
			continue
		}
		n -= locTable[i].weight
		if n <= 0 {
			if locTable[i].kind < 0 {
				break
			}
			var hid int
			if locTable[i].hidden == 2 {
				hid = g.rnd(0, 1)
			} else {
				hid = locTable[i].hidden
			}
			sl := g.createASubloc(row, col, hid, locTable[i].kind)
			s := g.nameGuild(locTable[i].kind)
			if s != "" {
				g.Subloc[sl].Name = s
			}
			break
		}
	}
}

// ---------------------------------------------------------------------------
// guild / sublocation naming
// ---------------------------------------------------------------------------

type guildName struct {
	skill  int
	weight int
	name   string
}

var guildNames = []guildName{
	{TerrStoneCir, 1, ""},
	{TerrGrove, 1, ""},
	{TerrBog, 1, ""},
	{TerrCave, 1, ""},

	{TerrGrave, 20, ""},
	{TerrGrave, 1, "Barrows"},
	{TerrGrave, 1, "Barrow Downs"},
	{TerrGrave, 1, "Barrow Hills"},
	{TerrGrave, 1, "Cairn Hills"},
	{TerrGrave, 1, "Catacombs"},
	{TerrGrave, 1, "Grave Mounds"},
	{TerrGrave, 1, "Place of the Dead"},
	{TerrGrave, 1, "Cemetary Hill"},
	{TerrGrave, 1, "Fields of Death"},

	{TerrRuins, 1, ""},

	{TerrBattlefield, 3, "Old battlefield"},
	{TerrBattlefield, 1, "Ancient battlefield"},
	{TerrBattlefield, 1, ""},

	{TerrEnchFor, 1, ""},
	{TerrRockyHill, 1, ""},
	{TerrTreeCir, 1, ""},
	{TerrPits, 1, "Cursed Pits"},
	{TerrPasture, 1, ""},
	{TerrPasture, 1, "Grassy field"},
	{TerrSacGrove, 1, ""},
	{TerrOasis, 1, ""},
	{TerrPopField, 1, ""},
	{TerrSandPit, 1, ""},
	{TerrYewGrove, 1, ""},

	{TerrTemple, 1, ""},
	{TerrLair, 1, ""},
}

func (g *Generator) nameGuild(skill int) string {
	sum := 0
	for i := range guildNames {
		if guildNames[i].skill == skill {
			sum += guildNames[i].weight
		}
	}
	if sum <= 0 {
		panic("name_guild: no matching skill")
	}

	n := g.rnd(1, sum)
	for i := range guildNames {
		if guildNames[i].skill == skill {
			n -= guildNames[i].weight
			if n <= 0 {
				return guildNames[i].name
			}
		}
	}
	panic("name_guild: unreachable")
}
