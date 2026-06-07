// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package store defines the G6 native JSON store entity types.
//
// A deliberate divergence from G3: G6 uses JSON as its native store because it
// is easier to work with in Go (see AGENTS.md). These types are the JSON-facing
// representation of the world entities the map generator produces and the
// engine loads, so the engine never has to parse the legacy flat-file format.
//
// The schema is intentionally human-readable: string kind/subkind, integer
// ids, snake_case fields, and booleans for flags.
package store

// Location is a "loc" entity. It covers all three flavors the generator emits:
// a province, a sublocation, or a region/continent. Fields that do not apply
// to a given flavor are omitted.
type Location struct {
	Id       int       `json:"id"`
	Kind     string    `json:"kind"`    // always "loc"
	Subkind  string    `json:"subkind"` // terrain name, or "region"
	Name     string    `json:"name,omitempty"`
	Where    int       `json:"where,omitempty"`     // container entity (the flat-file LI "wh")
	ProvDest *ProvDest `json:"prov_dest,omitempty"` // province exits N/E/S/W (provinces only)
	Here     []int     `json:"here,omitempty"`      // contents: roads, gates, sublocs (or provinces for a region)

	Hidden           int  `json:"hidden,omitempty"`
	SeaLane          bool `json:"sea_lane,omitempty"`
	SafeHaven        bool `json:"safe_haven,omitempty"`
	UldimFlag        int  `json:"uldim_flag,omitempty"`
	SummerbridgeFlag int  `json:"summerbridge_flag,omitempty"`
	MajorCity        int  `json:"major_city,omitempty"`
}

// ProvDest holds the four province exits, by compass direction. A zero value
// means "no exit that way." Mirrors the flat-file LO "pd N E S W" line.
type ProvDest struct {
	N int `json:"n"`
	E int `json:"e"`
	S int `json:"s"`
	W int `json:"w"`
}

// Gate is a "gate" entity: a one-way teleport from its containing location
// (Where) to a destination location (ToLoc).
type Gate struct {
	Id      int    `json:"id"`
	Kind    string `json:"kind"` // always "gate"
	Where   int    `json:"where"`
	ToLoc   int    `json:"to_loc"`
	SealKey int    `json:"seal_key,omitempty"`
}

// Road is a "road" entity: a connection (road, secret pass, channel, ...) from
// its containing location (Where) to a destination location (ToLoc).
type Road struct {
	Id     int    `json:"id"`
	Kind   string `json:"kind"` // always "road"
	Name   string `json:"name,omitempty"`
	Where  int    `json:"where"`
	ToLoc  int    `json:"to_loc"`
	Hidden int    `json:"hidden,omitempty"`
}
