package mapgen

import "fmt"

// ---------------------------------------------------------------------------
// loc file: provinces
// ---------------------------------------------------------------------------

func (g *Generator) printMap() {
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t == nil {
				continue
			}

			flag := true
			sl := false

			fmt.Fprintf(&g.loc, "%d loc %s\n", t.Region, TerrainNames[t.Terrain])

			if t.Name != "" && t.Name != "Unnamed" {
				fmt.Fprintf(&g.loc, "na %s\n", t.Name)
			}

			if t.UldimFlag != 0 {
				if !sl {
					fmt.Fprintf(&g.loc, "SL\n")
					sl = true
				}
				fmt.Fprintf(&g.loc, " uf %d\n", t.UldimFlag)
			}

			if t.SummerbridgeFlag != 0 {
				if !sl {
					fmt.Fprintf(&g.loc, "SL\n")
					sl = true
				}
				fmt.Fprintf(&g.loc, " sf %d\n", t.SummerbridgeFlag)
			}

			if t.SafeHaven != 0 {
				if !sl {
					fmt.Fprintf(&g.loc, "SL\n")
					sl = true
				}
				fmt.Fprintf(&g.loc, " sh 1\n")
			}

			if t.Inside != 0 {
				fmt.Fprintf(&g.loc, "LI\n")
				flag = false
				fmt.Fprintf(&g.loc, " wh %d\n", t.Inside+RegionOff)
			}

			g.printInsideSublocs(flag, row, col)

			fmt.Fprintf(&g.loc, "LO\n")
			fmt.Fprintf(&g.loc, " pd %d %d %d %d\n",
				g.provDest(t, DirN),
				g.provDest(t, DirE),
				g.provDest(t, DirS),
				g.provDest(t, DirW))

			if t.Hidden != 0 {
				fmt.Fprintf(&g.loc, " hi %d\n", t.Hidden)
			}
			if t.SeaLane != 0 {
				fmt.Fprintf(&g.loc, " sl 1\n")
			}

			fmt.Fprintf(&g.loc, "\n")
		}
	}
}

func (g *Generator) printInsideSublocs(flag bool, row, col int) {
	t := g.Map[row][col]
	count := 0

	for i := 0; i < len(t.Roads); i++ {
		count++
		if count == 1 {
			if flag {
				fmt.Fprintf(&g.loc, "LI\n")
			}
			fmt.Fprintf(&g.loc, " hl ")
		}
		if count%11 == 10 {
			fmt.Fprintf(&g.loc, "\\\n\t")
		}
		fmt.Fprintf(&g.loc, "%d ", t.Roads[i].EntNum)
	}

	for i := 0; i < len(t.GatesNum); i++ {
		count++
		if count == 1 {
			if flag {
				fmt.Fprintf(&g.loc, "LI\n")
			}
			fmt.Fprintf(&g.loc, " hl ")
		}
		if count%11 == 10 {
			fmt.Fprintf(&g.loc, "\\\n\t")
		}
		fmt.Fprintf(&g.loc, "%d ", t.GatesNum[i])
	}

	for i := 0; i < len(t.Subs); i++ {
		count++
		if count == 1 {
			if flag {
				fmt.Fprintf(&g.loc, "LI\n")
			}
			fmt.Fprintf(&g.loc, " hl ")
		}
		if count%11 == 10 {
			fmt.Fprintf(&g.loc, "\\\n\t")
		}
		fmt.Fprintf(&g.loc, "%d ", t.Subs[i])
	}

	if count != 0 {
		fmt.Fprintf(&g.loc, "\n")
	}
}

// ---------------------------------------------------------------------------
// loc file: sublocations
// ---------------------------------------------------------------------------

func (g *Generator) printSublocs() {
	for i := 1; i <= g.TopSubloc; i++ {
		s := g.Subloc[i]
		sl := false

		fmt.Fprintf(&g.loc, "%d loc %s\n", s.Region, TerrainNames[s.Terrain])

		if s.Name != "" && s.Name != "Unnamed" {
			fmt.Fprintf(&g.loc, "na %s\n", s.Name)
		}

		if s.Inside == 0 {
			panic("print_sublocs: subloc without inside")
		}
		fmt.Fprintf(&g.loc, "LI\n")
		fmt.Fprintf(&g.loc, " wh %d\n", s.Inside)
		g.printSublocGates(i)

		fmt.Fprintf(&g.loc, "LO\n")

		if s.Hidden != 0 {
			fmt.Fprintf(&g.loc, " hi %d\n", s.Hidden)
		}

		if s.MajorCity != 0 {
			if !sl {
				fmt.Fprintf(&g.loc, "SL\n")
				sl = true
			}
			fmt.Fprintf(&g.loc, " mc %d\n", s.MajorCity)
		}

		if s.SafeHaven != 0 {
			if !sl {
				fmt.Fprintf(&g.loc, "SL\n")
				sl = true
			}
			fmt.Fprintf(&g.loc, " sh 1\n")
		}

		fmt.Fprintf(&g.loc, "\n")
	}
}

func (g *Generator) printSublocGates(n int) {
	s := g.Subloc[n]
	count := 0

	for i := 0; i < len(s.Roads); i++ {
		count++
		if count == 1 {
			fmt.Fprintf(&g.loc, " hl ")
		}
		if count%11 == 10 {
			fmt.Fprintf(&g.loc, "\\\n\t")
		}
		fmt.Fprintf(&g.loc, "%d ", s.Roads[i].EntNum)
	}

	for i := 0; i < len(s.GatesNum); i++ {
		count++
		if count == 1 {
			fmt.Fprintf(&g.loc, " hl ")
		}
		if count%11 == 10 {
			fmt.Fprintf(&g.loc, "\\\n\t")
		}
		fmt.Fprintf(&g.loc, "%d ", s.GatesNum[i])
	}

	for i := 0; i < len(s.Subs); i++ {
		count++
		if count == 1 {
			fmt.Fprintf(&g.loc, " hl ")
		}
		if count%11 == 10 {
			fmt.Fprintf(&g.loc, "\\\n\t")
		}
		fmt.Fprintf(&g.loc, "%d ", s.Subs[i])
	}

	if count != 0 {
		fmt.Fprintf(&g.loc, "\n")
	}
}

// ---------------------------------------------------------------------------
// loc file: regions/continents
// ---------------------------------------------------------------------------

func (g *Generator) dumpContinents() {
	for i := 1; i <= g.InsideTop; i++ {
		fmt.Fprintf(&g.loc, "%d loc region\n", RegionOff+i)
		if g.InsideNames[i] != "" {
			fmt.Fprintf(&g.loc, "na %s\n", g.InsideNames[i])
		}
		g.printInsideLocs(i)
		fmt.Fprintf(&g.loc, "\n")
	}
}

func (g *Generator) printInsideLocs(n int) {
	count := 0
	for i := 0; i < len(g.InsideList[n]); i++ {
		count++
		if count == 1 {
			fmt.Fprintf(&g.loc, "LI\n")
			fmt.Fprintf(&g.loc, " hl ")
		}
		if count%11 == 10 {
			fmt.Fprintf(&g.loc, "\\\n\t")
		}
		fmt.Fprintf(&g.loc, "%d ", g.InsideList[n][i].Region)
	}
	if count != 0 {
		fmt.Fprintf(&g.loc, "\n")
	}
}

// ---------------------------------------------------------------------------
// road file
// ---------------------------------------------------------------------------

func (g *Generator) dumpRoads() {
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t == nil {
				continue
			}
			for j := 0; j < len(t.Roads); j++ {
				r := t.Roads[j]
				fmt.Fprintf(&g.road, "%d road 0\n", r.EntNum)
				if r.Name != "" {
					fmt.Fprintf(&g.road, "na %s\n", r.Name)
				}
				fmt.Fprintf(&g.road, "LI\n")
				fmt.Fprintf(&g.road, " wh %d\n", t.Region)
				fmt.Fprintf(&g.road, "GA\n")
				fmt.Fprintf(&g.road, " tl %d\n", r.ToLoc)
				if r.Hidden != 0 {
					fmt.Fprintf(&g.road, " rh %d\n", r.Hidden)
				}
				fmt.Fprintf(&g.road, "\n")
			}
		}
	}

	for i := 1; i <= g.TopSubloc; i++ {
		s := g.Subloc[i]
		for j := 0; j < len(s.Roads); j++ {
			r := s.Roads[j]
			fmt.Fprintf(&g.road, "%d road 0\n", r.EntNum)
			if r.Name != "" {
				fmt.Fprintf(&g.road, "na %s\n", r.Name)
			}
			fmt.Fprintf(&g.road, "LI\n")
			fmt.Fprintf(&g.road, " wh %d\n", s.Region)
			fmt.Fprintf(&g.road, "GA\n")
			fmt.Fprintf(&g.road, " tl %d\n", r.ToLoc)
			if r.Hidden != 0 {
				fmt.Fprintf(&g.road, " rh %d\n", r.Hidden)
			}
			fmt.Fprintf(&g.road, "\n")
		}
	}
}

// ---------------------------------------------------------------------------
// gate file
// ---------------------------------------------------------------------------

func (g *Generator) dumpGates() {
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t == nil {
				continue
			}
			for j := 0; j < len(t.GatesDest); j++ {
				fmt.Fprintf(&g.gate, "%d gate 0\n", t.GatesNum[j])
				fmt.Fprintf(&g.gate, "LI\n")
				fmt.Fprintf(&g.gate, " wh %d\n", t.Region)
				fmt.Fprintf(&g.gate, "GA\n")
				fmt.Fprintf(&g.gate, " tl %d\n", t.GatesDest[j])
				if t.GatesKey[j] != 0 {
					fmt.Fprintf(&g.gate, " sk %d\n", t.GatesKey[j])
				}
				fmt.Fprintf(&g.gate, "\n")
			}
		}
	}

	for i := 1; i <= g.TopSubloc; i++ {
		s := g.Subloc[i]
		for j := 0; j < len(s.GatesNum); j++ {
			fmt.Fprintf(&g.gate, "%d gate 0\n", s.GatesNum[j])
			fmt.Fprintf(&g.gate, "LI\n")
			fmt.Fprintf(&g.gate, " wh %d\n", s.Region)
			fmt.Fprintf(&g.gate, "GA\n")
			fmt.Fprintf(&g.gate, " tl %d\n", s.GatesDest[j])
			if s.GatesKey[j] != 0 {
				fmt.Fprintf(&g.gate, " sk %d\n", s.GatesKey[j])
			}
			fmt.Fprintf(&g.gate, "\n")
		}
	}
}
