package archaeology

import (
	"fmt"
	"sort"
	"strings"
)

type LayerObservation struct {
	UnitID, Colour, Texture, Inclusion string
	Thickness, Elevation               float64
}
type LayerProfile struct {
	Units                      []LayerObservation
	MinElevation, MaxElevation float64
}
type ProfileWarning struct {
	UnitID, Code, Message string
	Severity              string
}

func BuildProfile(units []LayerObservation) (LayerProfile, error) {
	if len(units) == 0 {
		return LayerProfile{}, fmt.Errorf("empty stratigraphic profile")
	}
	profile := LayerProfile{Units: append([]LayerObservation(nil), units...), MinElevation: units[0].Elevation, MaxElevation: units[0].Elevation}
	for _, u := range units {
		if u.UnitID == "" || strings.TrimSpace(u.Texture) == "" {
			return profile, fmt.Errorf("unit texture and id are required")
		}
		if u.Thickness <= 0 {
			return profile, fmt.Errorf("unit %s has non-positive thickness", u.UnitID)
		}
		if u.Elevation < profile.MinElevation {
			profile.MinElevation = u.Elevation
		}
		if u.Elevation > profile.MaxElevation {
			profile.MaxElevation = u.Elevation
		}
	}
	sort.Slice(profile.Units, func(i, j int) bool { return profile.Units[i].Elevation > profile.Units[j].Elevation })
	return profile, nil
}
func CheckProfile(p LayerProfile) []ProfileWarning {
	warnings := []ProfileWarning{}
	for i, u := range p.Units {
		if i > 0 {
			above := p.Units[i-1]
			expected := above.Elevation - above.Thickness
			if mathAbs(expected-u.Elevation) > 0.2 {
				warnings = append(warnings, ProfileWarning{UnitID: u.UnitID, Code: "gap", Message: "elevation does not meet adjacent layer", Severity: "warning"})
			}
		}
		if u.Thickness > 2 {
			warnings = append(warnings, ProfileWarning{UnitID: u.UnitID, Code: "thick", Message: "layer thickness needs a section drawing", Severity: "info"})
		}
		if strings.TrimSpace(u.Inclusion) == "" {
			warnings = append(warnings, ProfileWarning{UnitID: u.UnitID, Code: "inclusion-missing", Message: "inclusion description is empty", Severity: "warning"})
		}
	}
	return warnings
}
func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
func MergeProfiles(a, b LayerProfile) (LayerProfile, error) {
	if len(a.Units) == 0 {
		return b, nil
	}
	if len(b.Units) == 0 {
		return a, nil
	}
	combined := append(append([]LayerObservation(nil), a.Units...), b.Units...)
	seen := map[string]bool{}
	for _, u := range combined {
		if seen[u.UnitID] {
			return LayerProfile{}, fmt.Errorf("duplicate unit %s", u.UnitID)
		}
		seen[u.UnitID] = true
	}
	return BuildProfile(combined)
}
func ProfileRange(p LayerProfile) float64 { return p.MaxElevation - p.MinElevation }
func FindAtElevation(p LayerProfile, elevation float64) (LayerObservation, bool) {
	best := LayerObservation{}
	distance := 1e99
	for _, u := range p.Units {
		d := mathAbs(u.Elevation - elevation)
		if d < distance {
			distance = d
			best = u
		}
	}
	return best, distance < best.Thickness
}
