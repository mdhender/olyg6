package mapgen

// Map dimensions and entity-numbering constants, ported verbatim from the
// legacy C map generator (mapgen.c).
const (
	LineLen = 256

	MaxRow = 100
	MaxCol = 100

	// RegionOff is where region (continent/ocean) entity numbers start.
	RegionOff = 58760

	// MaxBox is the size of the entity allocation bitmap.
	MaxBox = 100000

	// MaxInside is the maximum number of continents/regions.
	MaxInside = 200

	// MaxSubloc is the maximum number of sublocations.
	MaxSubloc = 20000

	// Sublocation and city entity-number ranges.
	SublocLow  = 59000
	SublocHigh = 78999
	CityLow    = 56760
	CityHigh   = 58759
)

// Terrain types. The numeric values index into TerrainNames and are written
// directly into the generated database, so they must not change.
const (
	TerrLand        = 1
	TerrOcean       = 2
	TerrForest      = 3
	TerrSwamp       = 4
	TerrMountain    = 5
	TerrPlain       = 6
	TerrDesert      = 7
	TerrWater       = 8
	TerrIsland      = 9
	TerrStoneCir    = 10 // circle of stones
	TerrGrove       = 11 // mallorn grove
	TerrBog         = 12
	TerrCave        = 13
	TerrCity        = 14
	TerrGuild       = 15
	TerrGrave       = 16
	TerrRuins       = 17
	TerrBattlefield = 18
	TerrEnchFor     = 19 // enchanted forest
	TerrRockyHill   = 20
	TerrTreeCir     = 21
	TerrPits        = 22
	TerrPasture     = 23
	TerrOasis       = 24
	TerrYewGrove    = 25
	TerrSandPit     = 26
	TerrSacGrove    = 27 // sacred grove
	TerrPopField    = 28 // poppy field
	TerrTemple      = 29
	TerrLair        = 30 // dragon lair
)

// TerrainNames maps a terrain type to the string written to the database.
// Index 0 is the "<null>" placeholder, matching the C terr_s[] table.
var TerrainNames = []string{
	"<null>",
	"land",
	"ocean",
	"forest",
	"swamp",
	"mountain",
	"plain",
	"desert",
	"water",
	"island",
	"ring of stones",
	"mallorn grove",
	"bog",
	"cave",
	"city",
	"guild",
	"graveyard",
	"ruins",
	"battlefield",
	"enchanted forest",
	"rocky hill",
	"circle of trees",
	"pits",
	"pasture",
	"oasis",
	"yew grove",
	"sand pit",
	"sacred grove",
	"poppy field",
	"temple",
	"lair",
}

// Compass directions used by the adjacency helpers.
const (
	DirN  = 1
	DirE  = 2
	DirS  = 3
	DirW  = 4
	DirNE = 5
	DirSE = 6
	DirSW = 7
	DirNW = 8

	MaxDir = 9
)

// Road is a one-way connection (secret pass, channel, etc.) from a location
// to a destination location entity.
type Road struct {
	EntNum int
	Name   string
	ToLoc  int
	Hidden int
}

// Tile represents a single map province or sublocation. It mirrors the C
// "struct tile". Fields used purely as booleans in the original (SafeHaven,
// SeaLane, RegionBoundary) are kept as ints so their stored values match.
type Tile struct {
	SaveChar byte
	Region   int
	Name     string
	Terrain  int
	Hidden   int
	City     int
	Mark     int
	Inside   int
	Color    int // map coloring
	Row, Col int
	Depth    int

	SafeHaven        int
	SeaLane          int
	UldimFlag        int
	SummerbridgeFlag int
	RegionBoundary   int
	MajorCity        int

	Subs []int // sublocation entity numbers inside this tile

	GatesDest []int // gate destinations from here
	GatesNum  []int // gate entity numbers from here
	GatesKey  []int // gate keys

	Roads []*Road
}
