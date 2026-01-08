package solver

import "container/list"

// Edge represents a residual edge.
type Edge struct {
	To   int
	Rev  int
	Cap  int
	Flow int
}

type Flow struct {
	G [][]Edge
}

func NewFlow(n int) *Flow {
	g := make([][]Edge, n)
	return &Flow{G: g}
}

func (f *Flow) AddEdge(u, v, cap int) {
	forward := Edge{To: v, Rev: len(f.G[v]), Cap: cap, Flow: 0}
	back := Edge{To: u, Rev: len(f.G[u]), Cap: 0, Flow: 0}
	f.G[u] = append(f.G[u], forward)
	f.G[v] = append(f.G[v], back)
}

func (f *Flow) EdmondsKarp(s, t int) int {
	n := len(f.G)
	flow := 0

	for {
		prevV := make([]int, n)
		prevE := make([]int, n)
		for i := range prevV {
			prevV[i] = -1
			prevE[i] = -1
		}
		prevV[s] = s

		q := list.New()
		q.PushBack(s)

		for q.Len() > 0 && prevV[t] == -1 {
			v := q.Remove(q.Front()).(int)
			for ei, e := range f.G[v] {
				if prevV[e.To] != -1 {
					continue
				}
				if e.Cap-e.Flow <= 0 {
					continue
				}
				prevV[e.To] = v
				prevE[e.To] = ei
				q.PushBack(e.To)
				if e.To == t {
					break
				}
			}
		}

		if prevV[t] == -1 {
			break
		}

		// Find bottleneck
		add := int(^uint(0) >> 1)
		for v := t; v != s; v = prevV[v] {
			u := prevV[v]
			ei := prevE[v]
			e := f.G[u][ei]
			rem := e.Cap - e.Flow
			if rem < add {
				add = rem
			}
		}

		// Augment
		for v := t; v != s; v = prevV[v] {
			u := prevV[v]
			ei := prevE[v]
			rev := f.G[u][ei].Rev

			f.G[u][ei].Flow += add
			f.G[v][rev].Flow -= add
		}

		flow += add
	}

	return flow
}
