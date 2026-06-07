package mapgen

func (g *Generator) addRoad(from *Tile, toLoc, hidden int, name string) {
	r := &Road{
		EntNum: g.rndAllocNum(SublocLow, SublocHigh),
		ToLoc:  toLoc,
		Hidden: hidden,
		Name:   name,
	}
	from.Roads = append(from.Roads, r)
}

func (g *Generator) linkRoads(from, to *Tile, hidden int, name string) {
	// If there is a sublocation at an endpoint of the secret road, move the
	// road to come from the sublocation instead of the province.
	for i := 1; i <= g.TopSubloc; i++ {
		if g.Subloc[i].Inside == from.Region && g.Subloc[i].Terrain != TerrCity {
			from = g.Subloc[i]
			break
		}
	}
	for i := 1; i <= g.TopSubloc; i++ {
		if g.Subloc[i].Inside == to.Region && g.Subloc[i].Terrain != TerrCity {
			to = g.Subloc[i]
			break
		}
	}

	g.addRoad(from, to.Region, hidden, name)
	g.addRoad(to, from.Region, hidden, name)
}

func (g *Generator) makeRoads() {
	g.clearProvinceMarks()
	g.bridgeMapHoles()
	g.bridgeCaddyCorners()
	g.bridgeMountainPorts()
}

var bridgeDirS = []string{
	"-invalid-",
	"  n-s",
	"  e-w",
	"ne-sw",
	"nw-se",
}

func (g *Generator) bridgeMapHoles() {
	for row := 0; row < g.MaxRowUsed; row++ {
		for col := 0; col < g.MaxColUsed; col++ {
			if g.Map[row][col] == nil {
				if n := g.bridgeMapHoleSup(row, col); n != 0 {
					g.logf("%s map hole bridge at (%d,%d)\n", bridgeDirS[n], row, col)
				}
			}
		}
	}
	g.logf("\n")
}

func (g *Generator) bridgeMapHoleSup(row, col int) int {
	n := g.adjacentTileSup(row, col, DirN)
	s := g.adjacentTileSup(row, col, DirS)
	e := g.adjacentTileSup(row, col, DirE)
	w := g.adjacentTileSup(row, col, DirW)
	nw := g.adjacentTileSup(row, col, DirNW)
	sw := g.adjacentTileSup(row, col, DirSW)
	ne := g.adjacentTileSup(row, col, DirNE)
	se := g.adjacentTileSup(row, col, DirSE)

	if n.Mark != 0 || s.Mark != 0 || e.Mark != 0 || w.Mark != 0 ||
		nw.Mark != 0 || sw.Mark != 0 || ne.Mark != 0 || se.Mark != 0 {
		return 0
	}

	var l []int

	if n != nil && s != nil && n.Terrain != TerrOcean && s.Terrain != TerrOcean &&
		g.Map[n.Row][n.Col].Mark+g.Map[s.Row][s.Col].Mark == 0 {
		l = append(l, 1)
	}
	if e != nil && w != nil && e.Terrain != TerrOcean && w.Terrain != TerrOcean &&
		g.Map[e.Row][e.Col].Mark+g.Map[w.Row][w.Col].Mark == 0 {
		l = append(l, 2)
	}
	if ne != nil && sw != nil && ne.Terrain != TerrOcean && sw.Terrain != TerrOcean &&
		g.Map[ne.Row][ne.Col].Mark+g.Map[sw.Row][sw.Col].Mark == 0 {
		l = append(l, 3)
	}
	if se != nil && nw != nil && se.Terrain != TerrOcean && nw.Terrain != TerrOcean &&
		g.Map[se.Row][se.Col].Mark+g.Map[nw.Row][nw.Col].Mark == 0 {
		l = append(l, 4)
	}

	i := len(l)
	if i <= 0 {
		return 0
	}

	var name string
	switch g.holeRoadNameCnt % 3 {
	case 0:
		name = "Secret pass"
	case 1:
		name = "Secret route"
	case 2:
		name = "Old road"
	}
	g.holeRoadNameCnt++

	if n != nil {
		n.Mark += g.rnd(0, 1)
	}
	if e != nil {
		e.Mark += g.rnd(0, 1)
	}
	if w != nil {
		w.Mark += g.rnd(0, 1)
	}
	if s != nil {
		s.Mark += g.rnd(0, 1)
	}
	if nw != nil {
		nw.Mark += g.rnd(0, 1)
	}
	if ne != nil {
		ne.Mark += g.rnd(0, 1)
	}
	if sw != nil {
		sw.Mark += g.rnd(0, 1)
	}
	if se != nil {
		se.Mark += g.rnd(0, 1)
	}

	i = g.rnd(0, i-1)

	switch l[i] {
	case 1:
		g.linkRoads(n, s, 1, name)
		n.Mark = 1
		s.Mark = 1
	case 2:
		g.linkRoads(e, w, 1, name)
		e.Mark = 1
		w.Mark = 1
	case 3:
		g.linkRoads(ne, sw, 1, name)
		ne.Mark = 1
		sw.Mark = 1
	case 4:
		g.linkRoads(se, nw, 1, name)
		se.Mark = 1
		nw.Mark = 1
	}

	return l[i]
}

func (g *Generator) bridgeCaddyCorners() {
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			if g.Map[row][col] != nil && g.Map[row][col].Terrain != TerrOcean &&
				g.rnd(1, 35) == 35 {
				g.bridgeCornerSup(row, col)
			}
		}
	}
}

func (g *Generator) bridgeCornerSup(row, col int) int {
	n := g.adjacentTileSup(row, col, DirN)
	s := g.adjacentTileSup(row, col, DirS)
	e := g.adjacentTileSup(row, col, DirE)
	w := g.adjacentTileSup(row, col, DirW)
	nw := g.adjacentTileSup(row, col, DirNW)
	sw := g.adjacentTileSup(row, col, DirSW)
	ne := g.adjacentTileSup(row, col, DirNE)
	se := g.adjacentTileSup(row, col, DirSE)

	if (n != nil && n.Mark != 0) || (s != nil && s.Mark != 0) ||
		(e != nil && e.Mark != 0) || (w != nil && w.Mark != 0) ||
		(nw != nil && nw.Mark != 0) || (sw != nil && sw.Mark != 0) ||
		(ne != nil && ne.Mark != 0) || (se != nil && se.Mark != 0) {
		return 0
	}

	var name string
	switch g.cornerRoadNameCnt % 3 {
	case 0:
		name = "Secret pass"
	case 1:
		name = "Secret route"
	case 2:
		name = "Old road"
	}
	g.cornerRoadNameCnt++

	var l []int
	if nw != nil && nw.Terrain != TerrOcean {
		l = append(l, 1)
	}
	if ne != nil && ne.Terrain != TerrOcean {
		l = append(l, 2)
	}
	if se != nil && se.Terrain != TerrOcean {
		l = append(l, 3)
	}
	if sw != nil && sw.Terrain != TerrOcean {
		l = append(l, 4)
	}

	i := len(l)
	if i <= 0 {
		return 0
	}

	if n != nil {
		n.Mark += g.rnd(0, 1)
	}
	if e != nil {
		e.Mark += g.rnd(0, 1)
	}
	if w != nil {
		w.Mark += g.rnd(0, 1)
	}
	if s != nil {
		s.Mark += g.rnd(0, 1)
	}
	if nw != nil {
		nw.Mark += g.rnd(0, 1)
	}
	if ne != nil {
		ne.Mark += g.rnd(0, 1)
	}
	if sw != nil {
		sw.Mark += g.rnd(0, 1)
	}
	if se != nil {
		se.Mark += g.rnd(0, 1)
	}

	i = g.rnd(0, i-1)

	self := g.Map[row][col]
	switch l[i] {
	case 1:
		g.linkRoads(self, nw, 1, name)
		self.Mark = 1
		nw.Mark = 1
	case 2:
		g.linkRoads(self, ne, 1, name)
		self.Mark = 1
		ne.Mark = 1
	case 3:
		g.linkRoads(self, se, 1, name)
		self.Mark = 1
		se.Mark = 1
	case 4:
		g.linkRoads(self, sw, 1, name)
		self.Mark = 1
		sw.Mark = 1
	}

	return l[i]
}

func (g *Generator) bridgeMountainSup(row, col int) {
	from := g.Map[row][col]
	to := g.adjacentTileWater(row, col)

	if to.Terrain != TerrOcean {
		panic("bridge_mountain_sup: to is not ocean")
	}

	var name string
	switch g.rnd(1, 3) {
	case 1:
		name = "Narrow channel"
	case 2:
		name = "Rocky channel"
	case 3:
		name = "Secret sea route"
	}

	g.addRoad(from, to.Region, 1, name)
	g.addRoad(to, from.Region, 1, name)

	g.logf("secret sea route at (%d,%d)\n", from.Row, from.Col)
}

func (g *Generator) bridgeMountainPorts() {
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			if g.Map[row][col] != nil &&
				g.Map[row][col].Terrain == TerrMountain &&
				g.isPortCity(row, col) &&
				g.rnd(1, 7) == 7 {
				g.bridgeMountainSup(row, col)
			}
		}
	}
}
