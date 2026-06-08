// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package mapgen

import (
	"encoding/json"

	"github.com/mdhender/olyg6/pkg/store"
)

// writeJSON emits the G6 native JSON store (loc.json, gate.json, road.json)
// alongside the legacy flat files. It carries exactly the same entity data as
// the flat writers in output.go; the flat output is left untouched so the
// golden parity test still holds.
func (g *Generator) writeJSON() error {
	if err := g.writeJSONFile("loc.json", g.buildLocations()); err != nil {
		return err
	}
	if err := g.writeJSONFile("road.json", g.buildRoads()); err != nil {
		return err
	}
	return g.writeJSONFile("gate.json", g.buildGates())
}

// writeJSONFile marshals v as indented JSON (matching the input JSON house
// style) with a trailing newline and writes it to the output directory.
func (g *Generator) writeJSONFile(name string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return g.writeFile(name, append(data, '\n'))
}

// hereList concatenates a tile's contents in the same order the flat-file "hl"
// list uses: roads, then gates, then sublocations.
func hereList(t *Tile) []int {
	here := make([]int, 0, len(t.Roads)+len(t.GatesNum)+len(t.Subs))
	for _, r := range t.Roads {
		here = append(here, r.EntNum)
	}
	here = append(here, t.GatesNum...)
	here = append(here, t.Subs...)
	if len(here) == 0 {
		return nil
	}
	return here
}

// buildLocations builds the loc entities: provinces (row-major), then
// sublocations, then regions/continents — mirroring printMap, printSublocs,
// and dumpContinents.
func (g *Generator) buildLocations() []store.Location {
	var locs []store.Location

	// provinces
	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t == nil {
				continue
			}
			loc := store.Location{
				Id:               t.Region,
				Kind:             "loc",
				Subkind:          TerrainNames[t.Terrain],
				Here:             hereList(t),
				Hidden:           t.Hidden,
				SeaLane:          t.SeaLane != 0,
				SafeHaven:        t.SafeHaven != 0,
				UldimFlag:        t.UldimFlag,
				SummerbridgeFlag: t.SummerbridgeFlag,
				ProvDest: &store.ProvDest{
					N: g.provDest(t, DirN),
					E: g.provDest(t, DirE),
					S: g.provDest(t, DirS),
					W: g.provDest(t, DirW),
				},
			}
			if t.Name != "" && t.Name != "Unnamed" {
				loc.Name = t.Name
			}
			if t.Inside != 0 {
				loc.Where = t.Inside + RegionOff
			}
			locs = append(locs, loc)
		}
	}

	// sublocations
	for i := 1; i <= g.TopSubloc; i++ {
		s := g.Subloc[i]
		loc := store.Location{
			Id:        s.Region,
			Kind:      "loc",
			Subkind:   TerrainNames[s.Terrain],
			Where:     s.Inside,
			Here:      hereList(s),
			Hidden:    s.Hidden,
			SafeHaven: s.SafeHaven != 0,
			MajorCity: s.MajorCity,
		}
		if s.Name != "" && s.Name != "Unnamed" {
			loc.Name = s.Name
		}
		locs = append(locs, loc)
	}

	// regions / continents
	for i := 1; i <= g.InsideTop; i++ {
		loc := store.Location{
			Id:      RegionOff + i,
			Kind:    "loc",
			Subkind: "region",
			Name:    g.InsideNames[i],
		}
		for _, t := range g.InsideList[i] {
			loc.Here = append(loc.Here, t.Region)
		}
		locs = append(locs, loc)
	}

	return locs
}

// buildRoads builds the road entities: from provinces (row-major), then from
// sublocations — mirroring dumpRoads.
func (g *Generator) buildRoads() []store.Road {
	var roads []store.Road

	emit := func(where int, list []*Road) {
		for _, r := range list {
			roads = append(roads, store.Road{
				Id:     r.EntNum,
				Kind:   "road",
				Name:   r.Name,
				Where:  where,
				ToLoc:  r.ToLoc,
				Hidden: r.Hidden,
			})
		}
	}

	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t == nil {
				continue
			}
			emit(t.Region, t.Roads)
		}
	}
	for i := 1; i <= g.TopSubloc; i++ {
		s := g.Subloc[i]
		emit(s.Region, s.Roads)
	}

	return roads
}

// buildGates builds the gate entities: from provinces (row-major), then from
// sublocations — mirroring dumpGates.
func (g *Generator) buildGates() []store.Gate {
	var gates []store.Gate

	emit := func(t *Tile) {
		for j := 0; j < len(t.GatesNum); j++ {
			gates = append(gates, store.Gate{
				Id:      t.GatesNum[j],
				Kind:    "gate",
				Where:   t.Region,
				ToLoc:   t.GatesDest[j],
				SealKey: t.GatesKey[j],
			})
		}
	}

	for row := 0; row < MaxRow; row++ {
		for col := 0; col < MaxCol; col++ {
			t := g.Map[row][col]
			if t == nil {
				continue
			}
			emit(t)
		}
	}
	for i := 1; i <= g.TopSubloc; i++ {
		emit(g.Subloc[i])
	}

	return gates
}
