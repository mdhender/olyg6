package mapgen

import "fmt"

// The functions in this file reproduce the C program's diagnostic output,
// which is written to stderr (here, g.Log). They do not draw any random
// numbers and so do not affect the generated files or the saved seed.

func (g *Generator) countTiles() {
	var count [1000]int

	for r := 0; r < MaxRow; r++ {
		for c := 0; c < MaxCol; c++ {
			if g.Map[r][c] != nil {
				count[g.Map[r][c].Terrain]++
			}
		}
	}
	for i := 1; i <= g.TopSubloc; i++ {
		count[g.Subloc[i].Terrain]++
	}

	for i := 1; i < len(TerrainNames); i++ {
		g.logf("%-30s %d\n", TerrainNames[i], count[i])
	}
}

func (g *Generator) countCities() {
	for i := 1; i <= g.TopSubloc; i++ {
		if g.Subloc[i].Terrain == TerrCity {
			row := g.Subloc[i].Row
			col := g.Subloc[i].Col
			ins := g.Map[row][col].Inside
			g.insideNumCities[ins]++
		}
	}
}

func (g *Generator) printContinent(i int) {
	p := g.InsideList[i][0]

	name := g.InsideNames[i]
	if name == "" {
		name = fmt.Sprintf("?? (%d,%d)", p.Row, p.Col)
	}

	coord := fmt.Sprintf("(%d,%d)", p.Row, p.Col)
	gates := fmt.Sprintf("%d/%d", g.insideGatesFrom[i], g.insideGatesTo[i])
	nprovs := fmt.Sprintf("%d", len(g.InsideList[i]))
	ncities := fmt.Sprintf("%d", g.insideNumCities[i])

	g.logf("%-25s  %8s  %6s  %7s  %s\n", name, coord, nprovs, ncities, gates)
}

func (g *Generator) countContinents() {
	g.logf("\nLand regions:\n\n")
	g.logf("%-25s  %8s  %6s  %7s  %s\n", "name", "coord", "nprovs", "ncities", "gates (out/in)")
	g.logf("%-25s  %8s  %6s  %7s  %s\n", "----", "-----", "------", "-------", "--------------")

	for i := 1; i <= g.InsideTop; i++ {
		if g.InsideList[i][0].Terrain != TerrOcean {
			g.printContinent(i)
		}
	}

	g.logf("\n\nOceans:\n\n")
	g.logf("%-25s  %8s  %6s  %7s  %s\n", "name", "coord", "nprovs", "ncities", "gates (out/in)")
	g.logf("%-25s  %8s  %6s  %7s  %s\n", "----", "-----", "------", "-------", "--------------")

	for i := 1; i <= g.InsideTop; i++ {
		if g.InsideList[i][0].Terrain == TerrOcean {
			g.printContinent(i)
		}
	}

	g.logf("\n\n%d continents\n", g.InsideTop)
	g.logf("%d land locs\n", g.LandCount)
	g.logf("%d water locs\n", g.WaterCount)
}

func (g *Generator) countSublocs() {
	var count [100]int

	g.logf("\nsubloc counts:\n")

	g.clearProvinceMarks()

	for i := 1; i <= g.TopSubloc; i++ {
		if g.Subloc[i].Terrain == TerrIsland {
			row := g.Subloc[i].Row
			col := g.Subloc[i].Col
			g.Map[row][col].Mark++
		}
	}

	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			if g.Map[row][col] != nil && g.Map[row][col].Terrain == TerrOcean {
				m := g.Map[row][col].Mark
				if m >= 0 && m < 100 {
					count[m]++
				}
			}
		}
	}

	last := -1
	for i := 99; i >= 0; i-- {
		if count[i] != 0 {
			last = i
			break
		}
	}

	for i := 0; i <= last; i++ {
		has := "locs have"
		plural := "s"
		if count[i] == 1 {
			has = "loc has"
		}
		if i == 1 {
			plural = " "
		}
		pct := 0
		if g.WaterCount != 0 {
			pct = count[i] * 100 / g.WaterCount
		}
		g.logf("%6d %s %d island%s (%d%%)\n", count[i], has, i, plural, pct)
	}
}

func (g *Generator) countSublocCoverage() {
	var count [100]int

	g.clearProvinceMarks()

	for i := 1; i <= g.TopSubloc; i++ {
		if g.Subloc[i].Depth == 3 {
			row := g.Subloc[i].Row
			col := g.Subloc[i].Col
			g.Map[row][col].Mark++

			if g.Map[row][col].Mark >= 5 {
				g.logf("(%d,%d) has %d sublocs (region %d)\n",
					row, col, g.Map[row][col].Mark, g.Map[row][col].Region)
			}
		}
	}

	g.logf("\nsubloc coverage:\n")

	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			if g.Map[row][col] != nil && g.Map[row][col].Terrain != TerrOcean {
				m := g.Map[row][col].Mark
				if m >= 0 && m < 100 {
					count[m]++
				}
			}
		}
	}

	last := -1
	for i := 99; i >= 0; i-- {
		if count[i] != 0 {
			last = i
			break
		}
	}

	for i := 0; i <= last; i++ {
		has := "locs have"
		plural := "s"
		if count[i] == 1 {
			has = "loc has"
		}
		if i == 1 {
			plural = " "
		}
		pct := 0
		if g.LandCount != 0 {
			pct = count[i] * 100 / g.LandCount
		}
		g.logf("%6d %s %d subloc%s (%d%%)\n", count[i], has, i, plural, pct)
	}
}
