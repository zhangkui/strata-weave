package archaeology

import (
	"fmt"
	"math"
	"sort"
)

type Point struct{ X, Y, Z float64 }
type Polygon struct {
	ID     string
	Points []Point
}
type BoundingBox struct{ Min, Max Point }
type SpatialObservation struct {
	UnitID   string
	Location Point
	Accuracy float64
}

func Distance(a, b Point) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
func HorizontalDistance(a, b Point) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}
func ValidatePoint(p Point) error {
	if math.IsNaN(p.X) || math.IsNaN(p.Y) || math.IsNaN(p.Z) {
		return fmt.Errorf("point contains NaN")
	}
	if math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) || math.IsInf(p.Z, 0) {
		return fmt.Errorf("point contains infinity")
	}
	return nil
}
func ValidatePolygon(p Polygon) error {
	if p.ID == "" || len(p.Points) < 3 {
		return fmt.Errorf("polygon requires id and three points")
	}
	for _, point := range p.Points {
		if e := ValidatePoint(point); e != nil {
			return e
		}
	}
	if math.Abs(Area(p)) < 0.000001 {
		return fmt.Errorf("polygon has zero area")
	}
	return nil
}
func Area(p Polygon) float64 {
	if len(p.Points) < 3 {
		return 0
	}
	sum := 0.0
	for i, a := range p.Points {
		b := p.Points[(i+1)%len(p.Points)]
		sum += a.X*b.Y - b.X*a.Y
	}
	return sum / 2
}
func Centroid(p Polygon) Point {
	if len(p.Points) == 0 {
		return Point{}
	}
	area := Area(p)
	if area == 0 {
		sum := Point{}
		for _, x := range p.Points {
			sum.X += x.X
			sum.Y += x.Y
			sum.Z += x.Z
		}
		n := float64(len(p.Points))
		return Point{sum.X / n, sum.Y / n, sum.Z / n}
	}
	cx, cy := 0.0, 0.0
	for i, a := range p.Points {
		b := p.Points[(i+1)%len(p.Points)]
		cross := a.X*b.Y - b.X*a.Y
		cx += (a.X + b.X) * cross
		cy += (a.Y + b.Y) * cross
	}
	return Point{X: cx / (6 * area), Y: cy / (6 * area)}
}
func Bounds(points []Point) (BoundingBox, error) {
	if len(points) == 0 {
		return BoundingBox{}, fmt.Errorf("empty point set")
	}
	for _, p := range points {
		if e := ValidatePoint(p); e != nil {
			return BoundingBox{}, e
		}
	}
	b := BoundingBox{Min: points[0], Max: points[0]}
	for _, p := range points {
		b.Min.X = math.Min(b.Min.X, p.X)
		b.Min.Y = math.Min(b.Min.Y, p.Y)
		b.Min.Z = math.Min(b.Min.Z, p.Z)
		b.Max.X = math.Max(b.Max.X, p.X)
		b.Max.Y = math.Max(b.Max.Y, p.Y)
		b.Max.Z = math.Max(b.Max.Z, p.Z)
	}
	return b, nil
}
func Contains(p Polygon, x Point) bool {
	inside := false
	for i, j := 0, len(p.Points)-1; i < len(p.Points); j, i = i, i+1 {
		a, b := p.Points[i], p.Points[j]
		intersect := ((a.Y > x.Y) != (b.Y > x.Y)) && (x.X < (b.X-a.X)*(x.Y-a.Y)/(b.Y-a.Y)+a.X)
		if intersect {
			inside = !inside
		}
	}
	return inside
}
func Nearest(points []SpatialObservation, target Point, limit int) []SpatialObservation {
	out := append([]SpatialObservation(nil), points...)
	sort.Slice(out, func(i, j int) bool { return Distance(out[i].Location, target) < Distance(out[j].Location, target) })
	if limit <= 0 || limit > len(out) {
		limit = len(out)
	}
	return out[:limit]
}
func Cluster(points []SpatialObservation, radius float64) [][]SpatialObservation {
	if radius <= 0 {
		return nil
	}
	remaining := append([]SpatialObservation(nil), points...)
	groups := [][]SpatialObservation{}
	for len(remaining) > 0 {
		seed := remaining[0]
		remaining = remaining[1:]
		group := []SpatialObservation{seed}
		changed := true
		for changed {
			changed = false
			keep := remaining[:0]
			for _, p := range remaining {
				near := false
				for _, member := range group {
					if HorizontalDistance(p.Location, member.Location) <= radius {
						near = true
						break
					}
				}
				if near {
					group = append(group, p)
					changed = true
				} else {
					keep = append(keep, p)
				}
			}
			remaining = keep
		}
		groups = append(groups, group)
	}
	return groups
}
