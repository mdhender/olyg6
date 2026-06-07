package mapgen

// ---------------------------------------------------------------------------
// marks
// ---------------------------------------------------------------------------

func (g *Generator) clearProvinceMarks() {
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			if g.Map[row][col] != nil {
				g.Map[row][col].Mark = 0
			}
		}
	}
}

func (g *Generator) clearSublocMarks() {
	for i := 1; i <= g.TopSubloc; i++ {
		g.Subloc[i].Mark = 0
	}
}

func (g *Generator) markBadLocs() {
	for i := 1; i <= g.InsideTop; i++ {
		if g.InsideNames[i] == "Impassable Mountains" {
			for j := 0; j < len(g.InsideList[i]); j++ {
				g.InsideList[i][j].Mark = 1
			}
		}
	}

	// don't put gates in locations where magic can't be used
	for r := 0; r <= g.MaxRowUsed; r++ {
		for c := 0; c < g.MaxColUsed; c++ {
			if g.Map[r][c] != nil && g.Map[r][c].SafeHaven != 0 {
				g.Map[r][c].Mark = 1
			}
		}
	}
}

// ---------------------------------------------------------------------------
// random province / island selection
// ---------------------------------------------------------------------------

func (g *Generator) randomProvince(terr int) (row, col int) {
	sum := 0
	if terr == 0 {
		for r := 0; r <= g.MaxRowUsed; r++ {
			for c := 0; c < g.MaxColUsed; c++ {
				if g.Map[r][c] != nil && g.Map[r][c].Terrain != TerrOcean &&
					g.Map[r][c].Mark == 0 {
					sum++
				}
			}
		}
	} else {
		for r := 0; r <= g.MaxRowUsed; r++ {
			for c := 0; c < g.MaxColUsed; c++ {
				if g.Map[r][c] != nil && g.Map[r][c].Terrain == terr &&
					g.Map[r][c].Mark == 0 {
					sum++
				}
			}
		}
	}

	n := g.rnd(1, sum)

	if terr == 0 {
		for r := 0; r <= g.MaxRowUsed; r++ {
			for c := 0; c < g.MaxColUsed; c++ {
				if g.Map[r][c] != nil && g.Map[r][c].Terrain != TerrOcean &&
					g.Map[r][c].Mark == 0 {
					n--
					if n <= 0 {
						g.Map[r][c].Mark = 1
						return r, c
					}
				}
			}
		}
	} else {
		for r := 0; r <= g.MaxRowUsed; r++ {
			for c := 0; c < g.MaxColUsed; c++ {
				if g.Map[r][c] != nil && g.Map[r][c].Terrain == terr &&
					g.Map[r][c].Mark == 0 {
					n--
					if n <= 0 {
						g.Map[r][c].Mark = 1
						return r, c
					}
				}
			}
		}
	}

	panic("random_province: not found")
}

func (g *Generator) randomIsland() int {
	var i int
	for {
		n := g.rnd(1, g.NumIslands)
		i = 1
		for i <= g.TopSubloc {
			if g.Subloc[i].Terrain == TerrIsland {
				n--
				if n <= 0 {
					break
				}
			}
			i++
		}
		if g.Subloc[i].Mark == 0 {
			break
		}
	}
	g.Subloc[i].Mark = 1
	return i
}

// ---------------------------------------------------------------------------
// gates
// ---------------------------------------------------------------------------

func (g *Generator) makeGates() {
	g.gateProvinceIslands((g.LandCount + 199) / 200)
	g.randomProvinceGates((g.LandCount + 199) / 200)
	g.gateContinentalTour()
	g.gateStoneCircles()

	g.gateLandRing((g.LandCount + 749) / 750)
	g.gateLinkIslands((g.NumIslands + 149) / 150) // disjoint
	g.gateLinkIslands(g.NumIslands / 450)         // might cross
}

func (g *Generator) newGate(from, to *Tile, key int) {
	gateNum := g.rndAllocNum(SublocLow, SublocHigh)

	from.GatesNum = append(from.GatesNum, gateNum)
	from.GatesDest = append(from.GatesDest, to.Region)
	from.GatesKey = append(from.GatesKey, key)

	g.insideGatesFrom[g.Map[from.Row][from.Col].Inside]++
	g.insideGatesTo[g.Map[to.Row][to.Col].Inside]++
}

func (g *Generator) randomProvinceGates(n int) {
	g.clearProvinceMarks()
	g.markBadLocs()

	for i := 0; i < n; i++ {
		r1, c1 := g.randomProvince(0)
		r2, c2 := g.randomProvince(0)

		// QUIRK (preserved on purpose): the original C code reads
		//
		//	random_province(&r1, &c1, 0);
		//	random_province(&r2, &c2, 0);
		//	new_gate(map[r1][c1], map[r1][c2], 0);
		//
		// The destination is map[r1][c2] -- the row from the *first*
		// draw combined with the column from the *second* draw -- which
		// is almost certainly a typo for map[r2][c2]. The result is that
		// the gate's destination province is whatever tile happens to
		// sit at (r1, c2), not the second randomly chosen province.
		//
		// Both random_province calls still run (and still mark their
		// chosen tiles), so the RNG stream and the marking side effects
		// are unaffected; only which destination tile is used differs.
		// We reproduce map[r1][c2] verbatim so the generated output
		// stays byte-for-byte identical to the legacy generator. r2 is
		// drawn and marked but deliberately not used as an index.
		_ = r2
		g.newGate(g.Map[r1][c1], g.Map[r1][c2], 0)
	}
}

func (g *Generator) gateProvinceIslands(times int) {
	g.clearProvinceMarks()
	g.markBadLocs()
	g.clearSublocMarks()

	for j := 1; j <= times; j++ {
		r1, c1 := g.randomProvince(0)
		isle := g.randomIsland()
		r2, c2 := g.randomProvince(0)

		g.newGate(g.Map[r1][c1], g.Subloc[isle], 0)
		g.newGate(g.Subloc[isle], g.Map[r2][c2], 0)
	}
}

func (g *Generator) randomTileFromEachRegion() []*Tile {
	var l []*Tile

	for i := 1; i <= g.InsideTop; i++ {
		if g.InsideList[i][0].Terrain == TerrOcean {
			continue
		}
		if g.InsideNames[i] == "Impassable Mountains" {
			continue
		}

		var j int
		for {
			j = g.rnd(0, len(g.InsideList[i])-1)
			if g.InsideList[i][j].SafeHaven == 0 {
				break
			}
		}
		l = append(l, g.InsideList[i][j])
	}

	g.tshuffle(l)
	return l
}

func (g *Generator) shiftTourEndpoints(l []*Tile) []*Tile {
	var other []*Tile

	for i := 0; i < len(l); i++ {
		p := g.adjacentTileTerr(l[i].Row, l[i].Col)
		if p == nil {
			p = l[i]
		}

		q := g.adjacentTileTerr(p.Row, p.Col)
		if q == l[i] { // doubled back, retry
			q = g.adjacentTileTerr(p.Row, p.Col)
		}

		if q == nil || q.Terrain == TerrOcean || q.SafeHaven != 0 {
			g.logf("couldn't shift tour (%d,%d)\n", l[i].Row, l[i].Col)
			other = append(other, l[i])
		} else {
			other = append(other, q)
		}
	}

	return other
}

func (g *Generator) gateContinentalTour() {
	l := g.randomTileFromEachRegion()
	m := g.shiftTourEndpoints(l)

	if len(l) != len(m) {
		panic("gate_continental_tour length mismatch")
	}

	g.logf("\nContinental gate tour:\n")

	i := 0
	for ; i < len(l)-1; i++ {
		g.logf("\t(%2d,%2d) -> (%2d,%2d)\n", l[i].Row, l[i].Col, m[i+1].Row, m[i+1].Col)
		g.newGate(l[i], m[i+1], 0)
	}

	g.logf("\t(%2d,%2d) -> (%2d,%2d)\n\n", l[i].Row, l[i].Col, m[0].Row, m[0].Col)
	key := g.rnd(111, 333)
	g.newGate(l[i], m[0], key)
}

func (g *Generator) gateLinkIslands(rings int) {
	g.clearSublocMarks()

	for j := 1; j <= rings; j++ {
		num := g.rnd(5, 10)

		first := g.randomIsland()
		n := first

		for i := 1; i < num; i++ {
			next := g.randomIsland()
			g.newGate(g.Subloc[n], g.Subloc[next], 0)
			n = next
		}

		g.newGate(g.Subloc[n], g.Subloc[first], 0)
	}
}

func (g *Generator) gateLandRing(rings int) {
	g.clearProvinceMarks()
	g.markBadLocs()

	for j := 1; j <= rings; j++ {
		num := g.rnd(5, 10)
		rFirst, cFirst := g.randomProvince(0)

		rN, cN := rFirst, cFirst

		for i := 1; i < num; i++ {
			rNext, cNext := g.randomProvince(0)
			g.newGate(g.Map[rN][cN], g.Map[rNext][cNext], 0)
			rN, cN = rNext, cNext
		}

		g.newGate(g.Map[rN][cN], g.Map[rFirst][cFirst], 0)
	}
}

func (g *Generator) chooseRandomStoneCircle(l []*Tile, avoid1, avoid2 *Tile) *Tile {
	var i int
	for {
		i = g.rnd(0, len(l)-1)
		if l[i] != avoid1 && l[i] != avoid2 {
			break
		}
	}
	return l[i]
}

func (g *Generator) gateStoneCircles() {
	l := g.randomTileFromEachRegion()
	var circs []*Tile

	g.logf("\nRing of stones:\n")

	for i := 0; i < len(l); i++ {
		n := g.createASubloc(l[i].Row, l[i].Col, 1, TerrStoneCir)
		circs = append(circs, g.Subloc[n])
		g.logf("\t(%2d,%2d) in %s\n", l[i].Row, l[i].Col, g.InsideNames[l[i].Inside])
	}

	for i := 0; i < len(circs); i++ {
		first := g.chooseRandomStoneCircle(circs, circs[i], nil)
		second := g.chooseRandomStoneCircle(circs, circs[i], first)

		k1 := g.rnd(111, 333)
		g.newGate(circs[i], first, k1)
		k2 := g.rnd(111, 333)
		g.newGate(circs[i], second, k2)
	}

	g.clearProvinceMarks()
	g.markBadLocs()

	for i := 0; i < len(circs); i++ {
		for j := 1; j <= 5; j++ {
			row, col := g.randomProvince(0)
			var key int
			if g.rnd(0, 1) != 0 {
				key = 0
			} else {
				key = g.rnd(111, 333)
			}
			g.newGate(circs[i], g.Map[row][col], key)
		}
	}
}
