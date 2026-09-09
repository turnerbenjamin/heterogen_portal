package queryModel

import (
	"fmt"
	"strconv"
	"strings"
)

// Point represents a geographic point with a GeoJSON-compatible structure.
type Point struct {
	Type string `json:"type"`

	//Coordinates [x, y] i.e [long, lat]
	Coordinates [2]float64 `json:"coordinates"`
}

func NewPoint(longitude float64, latitude float64) Point {
	return Point{
		Type:        "Point",
		Coordinates: [2]float64{longitude, latitude},
	}
}

// UnmarshalJSON parses a point from WKT point data.
func (p *Point) UnmarshalJSON(data []byte) error {
	strData := strings.Trim(string(data), "\"")
	if strData == "null" || strData == "" {
		return nil
	}

	if !strings.HasPrefix(strData, "POINT") {
		return fmt.Errorf("invalid json for point: %s", strData)
	}

	start := strings.Index(strData, "(")
	end := strings.Index(strData, ")")
	if start == -1 || end == -1 || end <= start {
		return fmt.Errorf("invalid WKT point format: %s", strData)
	}

	coordsStr := strings.TrimSpace(strData[start+1 : end])
	parts := strings.Fields(coordsStr)
	if len(parts) != 2 {
		return fmt.Errorf("expected 2 coordinates in WKT point, got %d", len(parts))
	}

	long, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid longitude value: %w", err)
	}

	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid latitude value: %w", err)
	}

	p.Type = "Point"
	p.Coordinates = [2]float64{long, lat}
	return nil
}

// String returns the string representation of a Point
func (p *Point) String() string {
	return fmt.Sprintf("POINT (%f %f)", p.Coordinates[0], p.Coordinates[1])
}
