package mapgen

import (
	"encoding/json"
	"os"
	"strconv"
)

func isWhite(c byte) bool { return c == ' ' || c == '\t' }

// splitField splits s into the leading non-whitespace token and the remainder
// after the following run of whitespace.
func splitField(s string) (token, rest string) {
	i := 0
	for i < len(s) && !isWhite(s[i]) {
		i++
	}
	token = s[:i]
	for i < len(s) && isWhite(s[i]) {
		i++
	}
	return token, s[i:]
}

func parseRowCol(token string) (row, col int) {
	if idx := indexByte(token, ','); idx >= 0 {
		row, _ = strconv.Atoi(token[:idx])
		col, _ = strconv.Atoi(token[idx+1:])
	}
	return row, col
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func (g *Generator) allocInside() int {
	g.InsideTop++
	if g.InsideTop >= MaxInside {
		panic("inside_top overflow")
	}
	return g.InsideTop
}

// ---------------------------------------------------------------------------
// regions
// ---------------------------------------------------------------------------

func (g *Generator) readRegions() error {
	if data, err := os.ReadFile(g.inPath(g.InputRegions)); err != nil {
		g.logf("can't read %s\n", g.inPath(g.InputRegions))
		return err
	} else if err = json.Unmarshal(data, &g.regions); err != nil {
		g.logf("can't unmarshal %s\n", g.inPath(g.InputRegions))
		return err
	}
	return nil
}

func (g *Generator) setRegions() error {
	landCount := 0
	waterCount := 0

	for _, region := range g.regions {
		if region == nil {
			continue
		}

		row, col, name := region.Row, region.Col, region.Name

		if g.Map[row][col].Inside != 0 {
			g.logf("collision between %s and %s at (%d,%d)\n",
				name, g.InsideNames[g.Map[row][col].Inside], row, col)
			continue
		}

		ins := g.allocInside()
		g.InsideNames[ins] = name

		if g.Map[row][col].Terrain == TerrOcean {
			waterCount++
			g.floodWaterInside(row, col, ins)
		} else {
			landCount++
			g.floodLandInside(row, col, ins)
		}
	}

	g.logf("set_regions: named %d land regions, %d water regions\n",
		landCount, waterCount)

	// locate unnamed regions
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			if g.Map[row][col] != nil && g.Map[row][col].Inside == 0 {
				ins := g.allocInside()
				if g.Map[row][col].Terrain == TerrOcean {
					n := g.floodWaterInside(row, col, ins)
					g.logf("\tunnamed sea at  %d,%d (%d locs)\n", row, col, n)
				} else {
					n := g.floodLandInside(row, col, ins)
					g.logf("\tunnamed land at %d,%d (%d locs)\n", row, col, n)
				}
			}
		}
	}

	// collect the list of provinces in each region
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t != nil && t.Inside != 0 {
				g.InsideList[t.Inside] = append(g.InsideList[t.Inside], t)
			}
		}
	}

	return nil
}
func (g *Generator) floodLandInside(row, col, ins int) int {
	count := 1
	g.Map[row][col].Inside = ins

	if g.Map[row][col].RegionBoundary != 0 {
		return count
	}

	for dir := 1; dir < MaxDir; dir++ {
		p := g.adjacentTileSup(row, col, dir)
		if p == nil || p.Terrain == TerrOcean {
			continue
		}
		if p.Inside == ins {
			continue
		}
		if p.Inside != 0 {
			g.logf("flood_land_inside(%d,%d,%s) error\n", row, col, g.InsideNames[ins])
			panic("flood_land_inside")
		}
		count += g.floodLandInside(p.Row, p.Col, ins)
	}
	return count
}

func (g *Generator) floodWaterInside(row, col, ins int) int {
	count := 1
	g.Map[row][col].Inside = ins

	for dir := 1; dir < MaxDir; dir++ {
		p := g.adjacentTileSup(row, col, dir)
		if p == nil || p.Color == -1 || p.Color != g.Map[row][col].Color {
			continue
		}
		if p.Inside == ins {
			continue
		}
		if p.Inside != 0 {
			g.logf("flood_water_inside(%d,%d,%s) error\n", row, col, g.InsideNames[ins])
			panic("flood_water_inside")
		}
		count += g.floodWaterInside(p.Row, p.Col, ins)
	}
	return count
}

// ---------------------------------------------------------------------------
// province clumps (named land areas)
// ---------------------------------------------------------------------------
func (g *Generator) readLands() error {
	if data, err := os.ReadFile(g.inPath(g.InputLands)); err != nil {
		g.logf("can't read %s\n", g.inPath(g.InputLands))
		return err
	} else if err = json.Unmarshal(data, &g.lands); err != nil {
		g.logf("can't unmarshal %s\n", g.inPath(g.InputLands))
		return err
	}
	return nil
}

func (g *Generator) setProvinceClumps() error {
	count := 0
	for _, land := range g.lands {
		if land == nil {
			continue
		}

		row, col := land.Row, land.Col

		typeTok, name := land.Glyph, land.Name
		var typ byte
		if len(typeTok) > 0 {
			typ = typeTok[0]
		}

		if g.Map[row][col].SaveChar != typ {
			g.logf("Land '%s' expects '%c' at (%d,%d), got '%c'\n",
				name, typ, row, col, g.Map[row][col].SaveChar)
		}

		if g.Map[row][col].Name != "" {
			g.logf("clump collision between %s and %s at (%d,%d)\n",
				name, g.Map[row][col].Name, row, col)
		}

		g.floodLandClumps(row, col, name)
		count++
	}

	g.logf("read_and_set_province_clumps: named %d areas\n", count)
	return nil
}

func (g *Generator) unnamedProvinceClumps() {
	g.logf("Unnamed areas\n\n")

	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t != nil && t.Name == "" {
				if t.Terrain == TerrOcean {
					continue
				}
				n := g.floodLandClumps(row, col, "Unnamed")
				if t.SaveChar != 'o' {
					g.logf("%d,%d\t%c\t%d unnamed\n", row, col, t.SaveChar, n)
				}
			}
		}
	}
	g.logf("\n")
}

func (g *Generator) floodLandClumps(row, col int, name string) int {
	count := 1
	g.Map[row][col].Name = name

	for dir := 1; dir < MaxDir; dir++ {
		p := g.adjacentTileSup(row, col, dir)
		if p == nil || p.Terrain == TerrOcean || p.Color == -1 ||
			p.Color != g.Map[row][col].Color {
			continue
		}
		if p.Name == name {
			continue
		}
		if p.Name != "" {
			g.logf("flood_land_clumps(%d,%d,%s) error\n", row, col, name)
			panic("flood_land_clumps")
		}
		count += g.floodLandClumps(p.Row, p.Col, name)
	}
	return count
}
