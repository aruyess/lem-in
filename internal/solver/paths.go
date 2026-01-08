package solver

import (
	"errors"
	"sort"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
)

func bfsLevels(g *graph.Graph, start string) map[string]int {
	const inf = int(^uint(0) >> 1)
	dist := make(map[string]int, len(g.Adj))
	for n := range g.Adj {
		dist[n] = inf
	}
	dist[start] = 0

	q := []string{start}
	for head := 0; head < len(q); head++ {
		u := q[head]
		for _, v := range g.Neighbors(u) {
			if dist[v] != inf {
				continue
			}
			dist[v] = dist[u] + 1
			q = append(q, v)
		}
	}
	return dist
}

func layeredDirectedAdj(g *graph.Graph, levels map[string]int) map[string][]string {
	out := make(map[string][]string, len(g.Adj))
	for u := range g.Adj {
		for v := range g.Adj[u] {
			if levels[v] == levels[u]+1 {
				out[u] = append(out[u], v)
			}
		}
	}
	return out
}
var ErrNoPath = errors.New("no path")

// Path is a start->end route in terms of original room names (includes start and end).
type Path struct {
	Rooms []string
}
type splitNode struct {
	in  int
	out int
}

func (p Path) Edges() int { return len(p.Rooms) - 1 }

// splitNode is a node-splitting helper for vertex-capacity flow:
// each room becomes (in -> out) with a capacity on that edge.

func FindBestPaths(in parser.Input) ([]Path, error) {
	// 1) Build original undirected graph.
	g := graph.New()
	for name := range in.Rooms {
		g.AddNode(name)
	}
	g.AddNode(in.Start)
	g.AddNode(in.End)

	for _, e := range in.Links {
		g.AddUndirectedEdge(e[0], e[1])
	}

	// 2) Find disjoint paths on FULL graph (do not restrict to shortest only).
	paths := maxVertexDisjointPaths(g, in.Start, in.End, in.Ants)
	if len(paths) == 0 {
		return nil, ErrNoPath
	}

	// 3) Deterministic ordering (your existing logic).
	startOrder := make(map[string]int)
	order := 0
	for _, e := range in.Links {
		a, b := e[0], e[1]
		if a == in.Start {
			if _, ok := startOrder[b]; !ok {
				startOrder[b] = order
				order++
			}
			continue
		}
		if b == in.Start {
			if _, ok := startOrder[a]; !ok {
				startOrder[a] = order
				order++
			}
		}
	}

	sort.Slice(paths, func(i, j int) bool {
		if paths[i].Edges() != paths[j].Edges() {
			return paths[i].Edges() < paths[j].Edges()
		}
		ni, nj := "", ""
		if len(paths[i].Rooms) > 1 {
			ni = paths[i].Rooms[1]
		}
		if len(paths[j].Rooms) > 1 {
			nj = paths[j].Rooms[1]
		}
		oi, iok := startOrder[ni]
		oj, jok := startOrder[nj]
		if iok && jok && oi != oj {
			return oi < oj
		}
		if iok != jok {
			return iok
		}
		ri := paths[i].Rooms
		rj := paths[j].Rooms
		for k := 0; k < len(ri) && k < len(rj); k++ {
			if ri[k] != rj[k] {
				return ri[k] < rj[k]
			}
		}
		return len(ri) < len(rj)
	})

	// 4) Choose K that minimizes makespan.
	bestK := 1
	best := makespan(in.Ants, paths[:1])
	for k := 2; k <= len(paths); k++ {
		ms := makespan(in.Ants, paths[:k])
		if ms < best {
			best = ms
			bestK = k
		}
	}
	return paths[:bestK], nil
}

// makespan estimates number of turns required to move all ants using the given paths,
// assuming greedy assignment (each next ant goes to path with minimal edges+currentLoad).
// For a path with E edges and A assigned ants, finish time is roughly (E-1)+A.
func makespan(ants int, paths []Path) int {
	if ants <= 0 || len(paths) == 0 {
		return 0
	}

	load := make([]int, len(paths))
	maxFinish := 0

	for a := 0; a < ants; a++ {
		bestI := 0
		bestScore := paths[0].Edges() + load[0]

		for i := 1; i < len(paths); i++ {
			score := paths[i].Edges() + load[i]
			if score < bestScore {
				bestScore = score
				bestI = i
			}
		}

		load[bestI]++
		finish := (paths[bestI].Edges() - 1) + load[bestI]
		if finish > maxFinish {
			maxFinish = finish
		}
	}
	return maxFinish
}

// maxVertexDisjointPaths returns up to maxPaths vertex-disjoint paths from start to end.
// Implementation:
// - Node-splitting: each room becomes (in -> out). Capacity 1 for normal rooms, INF for start/end.
// - Each tunnel a-b becomes directed edges a_out -> b_in (cap 1) and b_out -> a_in (cap 1).
// - Run Edmonds–Karp to compute max flow.
// - Repeatedly extract paths by following edges with positive flow, consuming 1 unit per extracted path.
func maxVertexDisjointPaths(g *graph.Graph, start, end string, maxPaths int) []Path {
	// Build a stable list of rooms.
	rooms := make([]string, 0, len(g.Adj))
	for r := range g.Adj {
		rooms = append(rooms, r)
	}
	sort.Strings(rooms)

	// Room name -> index.
	idx := make(map[string]int, len(rooms))
	for i, r := range rooms {
		idx[r] = i
	}

	// Create split nodes for each room.
	nodes := make([]splitNode, len(rooms))
	next := 0
	for i := range rooms {
		nodes[i] = splitNode{in: next, out: next + 1}
		next += 2
	}

	si := idx[start]
	ei := idx[end]

	f := NewFlow(next)
	const inf = 1 << 30

	// Add in->out edges with vertex capacity.
	for i, r := range rooms {
		cap := 1
		if r == start || r == end {
			cap = inf
		}
		f.AddEdge(nodes[i].in, nodes[i].out, cap)
	}

	// Add tunnel edges (directed both ways) in deterministic order.
for _, a := range rooms {
	nb := g.Adj[a]
	neighbors := make([]string, 0, len(nb))
	for b := range nb {
		neighbors = append(neighbors, b)
	}
	sort.Strings(neighbors)

	ia := idx[a]
	for _, b := range neighbors {
		ib := idx[b]
		f.AddEdge(nodes[ia].out, nodes[ib].in, 1)
	}
}

	// Source is start_out; sink is end_in.
	s := nodes[si].out
	t := nodes[ei].in

	_ = f.EdmondsKarp(s, t)

	// Extract paths by consuming unit flow.
	paths := []Path{}
	for len(paths) < maxPaths {
		p, ok := extractOnePath(f, s, t, rooms, nodes)
		if !ok {
			break
		}
		paths = append(paths, p)
	}
	return paths
}

// extractOnePath finds a single s->t path in the residual graph by following edges with Flow>0,
// then consumes 1 unit of flow along that path, and converts the split-node sequence back to room names.
func extractOnePath(f *Flow, s, t int, rooms []string, nodes []splitNode) (Path, bool) {
	// DFS on positive-flow edges to build parent pointers.
	stack := []int{s}
	parentV := map[int]int{s: s}
	parentE := make(map[int]int)

	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if v == t {
			break
		}

		for ei, e := range f.G[v] {
			if e.Flow <= 0 {
				continue
			}
			if _, ok := parentV[e.To]; ok {
				continue
			}
			parentV[e.To] = v
			parentE[e.To] = ei
			stack = append(stack, e.To)
		}
	}

	if _, ok := parentV[t]; !ok {
		return Path{}, false
	}

	// Consume 1 flow along the found path.
	for v := t; v != s; v = parentV[v] {
		u := parentV[v]
		ei := parentE[v]
		rev := f.G[u][ei].Rev

		f.G[u][ei].Flow -= 1
		f.G[v][rev].Flow += 1
	}

	// Build nodeIndex -> roomName maps.
	roomByIn := make(map[int]string, len(rooms))
	roomByOut := make(map[int]string, len(rooms))
	for i, r := range rooms {
		roomByIn[nodes[i].in] = r
		roomByOut[nodes[i].out] = r
	}

	// Reconstruct the node sequence from s to t using parent pointers.
	order := []int{}
	for v := t; ; v = parentV[v] {
		order = append(order, v)
		if v == s {
			break
		}
	}
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}

	// Convert node sequence to room sequence:
	// - start name from s (an "out" node)
	// - then every time we hit an "in" node, append that room name.
	seq := []string{}
	if name, ok := roomByOut[s]; ok {
		seq = append(seq, name)
	}
	for i := 1; i < len(order); i++ {
		if name, ok := roomByIn[order[i]]; ok {
			seq = append(seq, name)
		}
	}

	// Deduplicate consecutive duplicates (can happen around start/end split).
	clean := make([]string, 0, len(seq))
	for _, x := range seq {
		if len(clean) == 0 || clean[len(clean)-1] != x {
			clean = append(clean, x)
		}
	}

	if len(clean) < 2 {
		return Path{}, false
	}
	return Path{Rooms: clean}, true
}
