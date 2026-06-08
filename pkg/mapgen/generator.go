package mapgen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mdhender/olyg6/pkg/prng"
	"github.com/mdhender/olyg6/pkg/store"
)

// Options configures a Generator.
type Options struct {
	// InputDir is the directory containing the input files. Defaults to ".".
	InputDir     string
	InputCities  string
	InputLands   string
	InputMap     string
	InputRegions string
	InputSeed    string
	// OutputDir is the directory the output files (loc, gate, road,
	// randseed) are written to. Defaults to ".".
	OutputDir string
	// Log receives the diagnostic output the C program writes to stderr.
	// Defaults to os.Stderr. Use io.Discard to silence it.
	Log io.Writer
}

// Generator holds the entire state of a single map-generation run. It is the
// direct analogue of the global state in the legacy C program.
type Generator struct {
	RNG *prng.RNG

	InputDir     string
	InputCities  string
	InputLands   string
	InputMap     string
	InputRegions string
	InputSeed    string

	OutputDir string

	Log io.Writer

	allocFlag [MaxBox]bool
	dirVector [MaxDir]int

	InsideNames     [MaxInside]string
	InsideList      [MaxInside][]*Tile
	insideGatesTo   [MaxInside]int
	insideGatesFrom [MaxInside]int
	insideNumCities [MaxInside]int
	InsideTop       int

	MaxColUsed int
	MaxRowUsed int

	WaterCount int
	LandCount  int
	NumIslands int

	Map [MaxRow][MaxCol]*Tile

	Subloc    []*Tile // 1-indexed; index 0 unused
	TopSubloc int

	// input buffers
	lands []*struct {
		Name  string `json:"name"`
		Row   int    `json:"row"`
		Col   int    `json:"col"`
		Glyph string `json:"glyph"`
	}
	regions []*struct {
		Name string `json:"name"`
		Row  int    `json:"row"`
		Col  int    `json:"col"`
	}

	// output buffers
	loc  bytes.Buffer
	gate bytes.Buffer
	road bytes.Buffer

	// Cities file reader state (matches the C static FILE *fp)
	citiesReader *bufio.Reader
	citiesOpened bool
	citiesFailed bool

	// bridge_corner_sup / bridge_map_hole_sup road name counters
	cornerRoadNameCnt int
	holeRoadNameCnt   int
}

// New returns a Generator configured with opts.
func New(opts Options) *Generator {
	g := &Generator{
		RNG:          prng.NewRNG(),
		InputDir:     opts.InputDir,
		InputCities:  opts.InputCities,
		InputLands:   opts.InputLands,
		InputMap:     opts.InputMap,
		InputRegions: opts.InputRegions,
		InputSeed:    opts.InputSeed,
		OutputDir:    opts.OutputDir,
		Log:          opts.Log,
		Subloc:       make([]*Tile, MaxSubloc),
	}
	if g.InputDir == "" {
		g.InputDir = "."
	}
	if g.InputCities == "" {
		g.InputCities = "cities.txt"
	}
	if g.InputLands == "" {
		g.InputLands = "lands.json"
	}
	if g.InputMap == "" {
		g.InputMap = "ascii-map.txt"
	}
	if g.InputRegions == "" {
		g.InputRegions = "regions.json"
	}
	if g.InputSeed == "" {
		g.InputSeed = "randseed"
	}
	if g.OutputDir == "" {
		g.OutputDir = "."
	}
	if g.Log == nil {
		g.Log = os.Stderr
	}
	return g
}

func (g *Generator) logf(format string, a ...any) {
	_, _ = fmt.Fprintf(g.Log, format, a...)
}

func (g *Generator) rnd(low, high int) int { return g.RNG.Rnd(low, high) }

func (g *Generator) inPath(name string) string  { return filepath.Join(g.InputDir, name) }
func (g *Generator) outPath(name string) string { return filepath.Join(g.OutputDir, name) }

// Run executes the full map-generation pipeline and writes the loc, gate,
// road, and randseed output files. It mirrors the C main() function.
func (g *Generator) Run() error {
	g.dirAssert()

	if err := store.LoadSeed(g.RNG, g.inPath(g.InputSeed)); err != nil {
		g.logf("%s could not be opened.\n", g.inPath(g.InputSeed))
		return err
	}
	if err := g.readRegions(); err != nil {
		g.logf("readRegions: %v\n", err)
		return err
	}
	if err := g.readLands(); err != nil {
		g.logf("readLands: %v\n", err)
		return err
	}
	if err := g.readCities(); err != nil {
		g.logf("readCities: %v\n", err)
		return err
	}

	if err := g.readMap(); err != nil {
		return err
	}
	g.fixTerrainLand()
	if err := g.setRegions(); err != nil {
		return err
	}
	if err := g.setProvinceClumps(); err != nil {
		return err
	}
	g.unnamedProvinceClumps()
	g.makeIslands()
	g.makeGraveyards()
	g.placeSublocations()
	g.makeGates()
	g.makeRoads()

	g.printMap()
	g.printSublocs()
	g.dumpContinents()
	g.countCities()
	g.countContinents()
	g.countSublocs()
	g.countSublocCoverage()
	g.dumpRoads()
	g.dumpGates()

	if err := g.writeFile("loc", g.loc.Bytes()); err != nil {
		return err
	}
	if err := g.writeFile("road", g.road.Bytes()); err != nil {
		return err
	}
	if err := g.writeFile("gate", g.gate.Bytes()); err != nil {
		return err
	}

	// Additionally emit the G6 native JSON store (loc.json, gate.json,
	// road.json). This carries the same entities as the flat files above;
	// the flat output is unchanged so golden parity still holds.
	if err := g.writeJSON(); err != nil {
		return err
	}

	g.countTiles()
	g.logf("\nhighest province = %d\n\n", g.Map[g.MaxRowUsed][g.MaxColUsed].Region)

	return store.SaveSeed(g.RNG, g.OutputDir)
}

func (g *Generator) writeFile(name string, data []byte) error {
	return os.WriteFile(g.outPath(name), data, 0644)
}

// ---------------------------------------------------------------------------
// entity allocation
// ---------------------------------------------------------------------------

func (g *Generator) rndAllocNum(low, high int) int {
	n := g.rnd(low, high)

	for i := n; i <= high; i++ {
		if !g.allocFlag[i] {
			g.allocFlag[i] = true
			return i
		}
	}
	for i := low; i < n; i++ {
		if !g.allocFlag[i] {
			g.allocFlag[i] = true
			return i
		}
	}

	g.logf("rnd_alloc_num(%d,%d) failed\n", low, high)
	return -1
}

// ---------------------------------------------------------------------------
// region / coordinate helpers
// ---------------------------------------------------------------------------

func rcToRegion(row, col int) int {
	return 10000 + (row * 100) + col
}

func regionRow(where int) int { return (where / 100) % 100 }
func regionCol(where int) int { return where % 100 }

func (g *Generator) dirAssert() {
	if rcToRegion(1, 1) != 10101 || regionRow(10101) != 1 || regionCol(10101) != 1 {
		panic("dir_assert failed")
	}
	if rcToRegion(99, 99) != 19999 {
		panic("dir_assert failed")
	}
}

// ---------------------------------------------------------------------------
// adjacency
// ---------------------------------------------------------------------------

func (g *Generator) adjacentTileSup(row, col, dir int) *Tile {
	switch dir {
	case DirN:
		row--
	case DirNE:
		row--
		col++
	case DirE:
		col++
	case DirSE:
		row++
		col++
	case DirS:
		row++
	case DirSW:
		row++
		col--
	case DirW:
		col--
	case DirNW:
		row--
		col--
	default:
		panic(fmt.Sprintf("location_direction: bad dir %d", dir))
	}

	if col < 0 {
		col = g.MaxColUsed
	}
	if col > g.MaxColUsed {
		col = 0
	}

	if row < 0 || row > 99 || col < 0 || col > 99 {
		return nil
	}

	return g.Map[row][col]
}

func (g *Generator) provDest(t *Tile, dir int) int {
	row, col := t.Row, t.Col

	switch dir {
	case DirN:
		row--
	case DirE:
		col++
	case DirS:
		row++
	case DirW:
		col--
	default:
		panic(fmt.Sprintf("location_direction: bad dir %d", dir))
	}

	if row < 0 || row > 99 {
		return 0
	}
	if col < 0 {
		col = g.MaxColUsed
	}
	if col > g.MaxColUsed {
		col = 0
	}

	if g.Map[row][col] == nil {
		return 0
	}
	return g.Map[row][col].Region
}

func (g *Generator) randomizeDirVector() {
	g.dirVector[0] = 0
	for i := 1; i < MaxDir; i++ {
		g.dirVector[i] = i
	}
	for i := 1; i < MaxDir; i++ {
		swap := g.rnd(i, MaxDir-1)
		if i != swap {
			g.dirVector[i], g.dirVector[swap] = g.dirVector[swap], g.dirVector[i]
		}
	}
}

func (g *Generator) adjacentTileWater(row, col int) *Tile {
	var p *Tile
	g.randomizeDirVector()

	i := 1
	for !(p != nil && p.Terrain == TerrOcean) && i < MaxDir {
		p = g.adjacentTileSup(row, col, g.dirVector[i])
		i++
	}
	if i < MaxDir {
		return p
	}
	return nil
}

func (g *Generator) adjacentTileTerr(row, col int) *Tile {
	var p *Tile
	g.randomizeDirVector()

	i := 1
	for !(p != nil && p.Terrain != TerrLand && p.Terrain != TerrOcean) && i < MaxDir {
		p = g.adjacentTileSup(row, col, g.dirVector[i])
		i++
	}
	if i < MaxDir {
		return p
	}
	return nil
}

func (g *Generator) isPortCity(row, col int) bool {
	n := g.adjacentTileSup(row, col, DirN)
	s := g.adjacentTileSup(row, col, DirS)
	e := g.adjacentTileSup(row, col, DirE)
	w := g.adjacentTileSup(row, col, DirW)

	if (n != nil && n.Terrain == TerrOcean) ||
		(s != nil && s.Terrain == TerrOcean) ||
		(e != nil && e.Terrain == TerrOcean) ||
		(w != nil && w.Terrain == TerrOcean) {
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Cities file
// ---------------------------------------------------------------------------

func (g *Generator) readCities() error {
	f, err := os.Open(g.inPath(g.InputCities))
	if err != nil {
		g.logf("can't open %s: %v\n", g.inPath(g.InputCities), err)
		g.citiesFailed = true
		return err
	}
	g.citiesReader = bufio.NewReader(f)
	g.citiesFailed, g.citiesOpened = false, true
	return nil
}

// nextCityName mimics the C create_a_city() use of getlin(fp): it returns the
// next line of the Cities file, or ("", false) once the file is exhausted or
// cannot be opened.
func (g *Generator) nextCityName() (string, bool) {
	if g.citiesReader == nil {
		panic("assert(citiesReader != nil)")
	}
	if g.citiesFailed {
		return "", false
	}

	line, err := g.citiesReader.ReadString('\n')
	if line == "" && err != nil {
		return "", false
	}
	line = strings.TrimRight(line, "\n")
	return line, true
}

// ---------------------------------------------------------------------------
// subloc / city / building creation
// ---------------------------------------------------------------------------

func (g *Generator) createASubloc(row, col, hidden, kind int) int {
	g.TopSubloc++
	if g.TopSubloc >= MaxSubloc {
		panic("top_subloc overflow")
	}

	t := &Tile{}
	g.Subloc[g.TopSubloc] = t
	if kind == TerrCity {
		t.Region = g.rndAllocNum(CityLow, CityHigh)
	} else {
		t.Region = g.rndAllocNum(SublocLow, SublocHigh)
	}
	t.Inside = g.Map[row][col].Region
	t.Row = row
	t.Col = col
	t.Hidden = hidden
	t.Terrain = kind
	t.Depth = 3

	if kind == TerrCity {
		g.Map[row][col].City = 2
	}

	g.Map[row][col].Subs = append(g.Map[row][col].Subs, t.Region)

	return g.TopSubloc
}

func (g *Generator) createACity(row, col int, name string, hasName bool, major int) int {
	if !hasName {
		if n, ok := g.nextCityName(); ok {
			name = n
		}
	}

	n := g.createASubloc(row, col, 0, TerrCity)
	g.Subloc[n].Name = name
	g.Subloc[n].MajorCity = major
	return n
}

func (g *Generator) createAGraveyard(row, col int) {
	hidden := g.rnd(0, 1)
	n := g.createASubloc(row, col, hidden, TerrGrave)
	s := g.nameGuild(TerrGrave)
	if s != "" {
		g.Subloc[n].Name = s
	}
}

// ---------------------------------------------------------------------------
// map reading
// ---------------------------------------------------------------------------

func (g *Generator) readMap() error {
	f, err := os.Open(g.inPath(g.InputMap))
	if err != nil {
		return fmt.Errorf("can't open %s: %w", g.inPath(g.InputMap), err)
	}
	defer func() {
		_ = f.Close()
	}()

	r := bufio.NewReader(f)
	row := 0

	for {
		line, rerr := r.ReadString('\n')
		if line == "" {
			break
		}

		for col := 0; col < len(line) && line[col] != '\n'; col++ {
			ch := line[col]
			if ch == '#' { // hole in map
				continue
			}

			if row > g.MaxRowUsed {
				g.MaxRowUsed = row
			}
			if col > g.MaxColUsed {
				g.MaxColUsed = col
			}

			t := &Tile{
				Row:    row,
				Col:    col,
				Region: rcToRegion(row, col),
				Depth:  2,
			}
			g.Map[row][col] = t

			color := 0
			terrain := 0

			switch ch {
			case ';':
				t.SeaLane = 1
				terrain = TerrOcean
				color = 1
			case ',':
				terrain = TerrOcean
				color = 1
			case ':':
				t.SeaLane = 1
				terrain = TerrOcean
				color = 2
			case '.':
				terrain = TerrOcean
				color = 2
			case '~':
				t.SeaLane = 1
				terrain = TerrOcean
				color = 3
			case ' ':
				terrain = TerrOcean
				color = 3
			case '"':
				t.SeaLane = 1
				terrain = TerrOcean
				color = 4
			case '\'':
				terrain = TerrOcean
				color = 4
			case 'p':
				color = 5
				terrain = TerrPlain
			case 'P':
				color = 6
				terrain = TerrPlain
			case 'd':
				color = 7
				terrain = TerrDesert
			case 'D':
				color = 8
				terrain = TerrDesert
			case 'm':
				color = 9
				terrain = TerrMountain
			case 'M':
				color = 10
				terrain = TerrMountain
			case 's':
				color = 11
				terrain = TerrSwamp
			case 'S':
				color = 12
				terrain = TerrSwamp
			case 'f':
				color = 13
				terrain = TerrForest
			case 'F':
				color = 14
				terrain = TerrForest
			case 'o':
				switch g.rnd(1, 10) {
				case 1, 2, 3:
					terrain = TerrForest
				case 4, 5, 6:
					terrain = TerrPlain
				case 7, 8:
					terrain = TerrMountain
				case 9:
					terrain = TerrSwamp
				case 10:
					terrain = TerrDesert
				}
				color = -1
			case '^':
				color = 9
				terrain = TerrMountain
				t.UldimFlag = 1
				t.RegionBoundary = 1
			case 'v':
				color = 9
				terrain = TerrMountain
				t.UldimFlag = 2
				t.RegionBoundary = 1
			case '{':
				color = 16
				terrain = TerrMountain
				t.UldimFlag = 3
				t.Name = "Uldim pass"
				t.RegionBoundary = 1
			case '}':
				color = 16
				terrain = TerrMountain
				t.UldimFlag = 4
				t.Name = "Uldim pass"
				t.RegionBoundary = 1
			case ']':
				terrain = TerrSwamp
				t.SummerbridgeFlag = 1
				t.Name = "Summerbridge"
				t.RegionBoundary = 1
			case '[':
				terrain = TerrSwamp
				t.SummerbridgeFlag = 2
				t.Name = "Summerbridge"
				t.RegionBoundary = 1
			case 'O':
				terrain = TerrMountain
				color = -1
				t.Name = "Mt. Olympus"
			case '1':
				terrain = TerrForest
				color = 19
				t.SafeHaven = 1
				n := g.createACity(row, col, "Drassa", true, 1)
				g.Subloc[n].SafeHaven = 1
				g.logf("Start city #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '2':
				terrain = TerrForest
				color = 19
				t.SafeHaven = 1
				n := g.createACity(row, col, "Rimmon", true, 1)
				g.Subloc[n].SafeHaven = 1
				g.logf("Start city #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '3':
				terrain = TerrForest
				color = 19
				t.SafeHaven = 1
				n := g.createACity(row, col, "Harn", true, 1)
				g.Subloc[n].SafeHaven = 1
				g.logf("Start city #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '4':
				terrain = TerrForest
				color = 19
				t.SafeHaven = 1
				n := g.createACity(row, col, "Imperial City", true, 1)
				g.Subloc[n].SafeHaven = 1
				g.logf("Imperical City #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '5':
				terrain = TerrForest
				color = 19
				t.SafeHaven = 1
				n := g.createACity(row, col, "Port Aurnos", true, 1)
				g.Subloc[n].SafeHaven = 1
				g.logf("Start city #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '6':
				terrain = TerrForest
				color = 19
				t.SafeHaven = 1
				n := g.createACity(row, col, "Greyfell", true, 1)
				g.Subloc[n].SafeHaven = 1
				g.logf("Start city #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '7':
				terrain = TerrForest
				color = 19
				t.SafeHaven = 1
				n := g.createACity(row, col, "Yellowleaf", true, 1)
				g.Subloc[n].SafeHaven = 1
				g.logf("Start city #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '8':
				terrain = TerrForest
				color = 19
				n := g.createACity(row, col, "Golden City", true, 1)
				g.logf("Golden City #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '9', '0':
				terrain = TerrForest
				color = 19
				t.SafeHaven = 1
				n := g.createACity(row, col, "", false, 1)
				g.Subloc[n].SafeHaven = 1
				g.logf("Start city #%c %s at (%d,%d)\n", ch, g.Subloc[n].Name, row, col)
			case '*':
				terrain = TerrLand
				g.createACity(row, col, "", false, 1)
			case '%':
				terrain = TerrLand
				g.createACity(row, col, "", false, 0)
			default:
				g.logf("unknown terrain %c\n", ch)
				panic("read_map: unknown terrain")
			}

			t.SaveChar = ch
			t.Terrain = terrain
			t.Color = color

			if terrain == TerrWater || terrain == TerrOcean {
				g.WaterCount++
			} else {
				g.LandCount++
			}
		}

		row++
		if rerr != nil {
			break
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// terrain inference
// ---------------------------------------------------------------------------

func (g *Generator) fixTerrainLand() {
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t == nil || t.Terrain != TerrLand {
				continue
			}
			p := g.adjacentTileTerr(row, col)
			if p != nil && p.Terrain != TerrLand && p.Terrain != TerrOcean {
				t.Terrain = p.Terrain
				t.Color = p.Color
			} else {
				g.logf("fix_terrain: could not infer type of (%d,%d)\n", row, col)
				g.logf("    assuming 'forest'\n")
				t.Terrain = TerrForest
			}
		}
	}
}
