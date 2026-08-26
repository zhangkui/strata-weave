package archaeology

import (
	"fmt"
	"sort"
	"strings"
)

type FindDescriptor struct {
	CatalogueNo, Kind, Material, Condition, UnitID string
	WeightGrams                                    float64
	Dimensions                                     [3]float64
	Tags                                           []string
}
type Assemblage struct {
	UnitID      string
	Total       int
	ByMaterial  map[string]int
	ByKind      map[string]int
	TotalWeight float64
	Conditions  map[string]int
}
type Similarity struct {
	Left, Right string
	Score       float64
	SharedTags  []string
}

func ValidateDescriptor(f FindDescriptor) error {
	if f.CatalogueNo == "" || f.Kind == "" || f.UnitID == "" {
		return fmt.Errorf("catalogue number, kind and unit are required")
	}
	if f.WeightGrams < 0 {
		return fmt.Errorf("negative find weight")
	}
	for _, d := range f.Dimensions {
		if d < 0 {
			return fmt.Errorf("negative dimension")
		}
	}
	seen := map[string]bool{}
	for _, tag := range f.Tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			return fmt.Errorf("empty find tag")
		}
		if seen[tag] {
			return fmt.Errorf("repeated find tag %q", tag)
		}
		seen[tag] = true
	}
	return nil
}
func BuildAssemblage(unit string, finds []FindDescriptor) (Assemblage, error) {
	a := Assemblage{UnitID: unit, ByMaterial: map[string]int{}, ByKind: map[string]int{}, Conditions: map[string]int{}}
	for _, f := range finds {
		if e := ValidateDescriptor(f); e != nil {
			return a, e
		}
		if f.UnitID != unit {
			return a, fmt.Errorf("find %s belongs to a different unit", f.CatalogueNo)
		}
		a.Total++
		a.ByMaterial[normalize(f.Material)]++
		a.ByKind[normalize(f.Kind)]++
		a.Conditions[normalize(f.Condition)]++
		a.TotalWeight += f.WeightGrams
	}
	return a, nil
}
func CompareFinds(a, b FindDescriptor) Similarity {
	left := tagSet(a.Tags)
	right := tagSet(b.Tags)
	shared := []string{}
	for tag := range left {
		if right[tag] {
			shared = append(shared, tag)
		}
	}
	sort.Strings(shared)
	union := len(left)
	for tag := range right {
		if !left[tag] {
			union++
		}
	}
	score := 0.0
	if union > 0 {
		score = float64(len(shared)) / float64(union)
	}
	if normalize(a.Material) == normalize(b.Material) {
		score += 0.15
	}
	if normalize(a.Kind) == normalize(b.Kind) {
		score += 0.2
	}
	if score > 1 {
		score = 1
	}
	return Similarity{Left: a.CatalogueNo, Right: b.CatalogueNo, Score: score, SharedTags: shared}
}
func normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
func tagSet(tags []string) map[string]bool {
	out := map[string]bool{}
	for _, t := range tags {
		if t = normalize(t); t != "" {
			out[t] = true
		}
	}
	return out
}
