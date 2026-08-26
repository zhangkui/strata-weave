package archaeology

import (
	"fmt"
	"sort"
)

type Edge struct{ Earlier, Later string }
type Matrix struct {
	edges map[string]map[string]bool
	nodes map[string]bool
}

func NewMatrix(edges []Edge) (*Matrix, error) {
	m := &Matrix{edges: map[string]map[string]bool{}, nodes: map[string]bool{}}
	for _, e := range edges {
		if err := m.Add(e); err != nil {
			return nil, err
		}
	}
	return m, nil
}
func (m *Matrix) Add(e Edge) error {
	if e.Earlier == "" || e.Later == "" {
		return fmt.Errorf("matrix edge endpoint missing")
	}
	if e.Earlier == e.Later {
		return fmt.Errorf("self reference %s", e.Earlier)
	}
	if m.Reachable(e.Later, e.Earlier) {
		return fmt.Errorf("edge %s -> %s closes a cycle", e.Earlier, e.Later)
	}
	if m.edges[e.Earlier] == nil {
		m.edges[e.Earlier] = map[string]bool{}
	}
	m.edges[e.Earlier][e.Later] = true
	m.nodes[e.Earlier] = true
	m.nodes[e.Later] = true
	return nil
}
func (m *Matrix) Remove(e Edge) bool {
	if m.edges[e.Earlier] == nil || !m.edges[e.Earlier][e.Later] {
		return false
	}
	delete(m.edges[e.Earlier], e.Later)
	return true
}
func (m *Matrix) Reachable(from, to string) bool {
	seen := map[string]bool{}
	stack := []string{from}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n == to {
			return true
		}
		if seen[n] {
			continue
		}
		seen[n] = true
		for next := range m.edges[n] {
			stack = append(stack, next)
		}
	}
	return false
}
func (m *Matrix) EarlierThan(unit string) []string { return m.walkReverse(unit) }
func (m *Matrix) LaterThan(unit string) []string   { return m.walkForward(unit) }
func (m *Matrix) walkForward(unit string) []string {
	seen := map[string]bool{}
	stack := []string{unit}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for next := range m.edges[n] {
			if !seen[next] {
				seen[next] = true
				stack = append(stack, next)
			}
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
func (m *Matrix) walkReverse(unit string) []string {
	reverse := map[string][]string{}
	for from, targets := range m.edges {
		for to := range targets {
			reverse[to] = append(reverse[to], from)
		}
	}
	seen := map[string]bool{}
	stack := []string{unit}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, next := range reverse[n] {
			if !seen[next] {
				seen[next] = true
				stack = append(stack, next)
			}
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
func (m *Matrix) TopologicalOrder() ([]string, error) {
	incoming := map[string]int{}
	for n := range m.nodes {
		incoming[n] = 0
	}
	for _, targets := range m.edges {
		for to := range targets {
			incoming[to]++
		}
	}
	ready := []string{}
	for n, count := range incoming {
		if count == 0 {
			ready = append(ready, n)
		}
	}
	sort.Strings(ready)
	order := []string{}
	for len(ready) > 0 {
		n := ready[0]
		ready = ready[1:]
		order = append(order, n)
		for next := range m.edges[n] {
			incoming[next]--
			if incoming[next] == 0 {
				ready = append(ready, next)
			}
		}
		sort.Strings(ready)
	}
	if len(order) != len(m.nodes) {
		return nil, fmt.Errorf("matrix includes a cycle")
	}
	return order, nil
}
func (m *Matrix) Layers() ([][]string, error) {
	order, e := m.TopologicalOrder()
	if e != nil {
		return nil, e
	}
	level := map[string]int{}
	for _, n := range order {
		for next := range m.edges[n] {
			if level[next] < level[n]+1 {
				level[next] = level[n] + 1
			}
		}
	}
	max := 0
	for _, v := range level {
		if v > max {
			max = v
		}
	}
	layers := make([][]string, max+1)
	for _, n := range order {
		layers[level[n]] = append(layers[level[n]], n)
	}
	for i := range layers {
		sort.Strings(layers[i])
	}
	return layers, nil
}
func (m *Matrix) CriticalPath() ([]string, error) {
	order, e := m.TopologicalOrder()
	if e != nil {
		return nil, e
	}
	distance := map[string]int{}
	previous := map[string]string{}
	last := ""
	for _, n := range order {
		if last == "" || distance[n] > distance[last] {
			last = n
		}
		for next := range m.edges[n] {
			if distance[next] < distance[n]+1 {
				distance[next] = distance[n] + 1
				previous[next] = n
			}
		}
	}
	path := []string{}
	for n := last; n != ""; n = previous[n] {
		path = append(path, n)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, nil
}
