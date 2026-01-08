package graph

type Graph struct {
	Adj map[string]map[string]bool
}

func New() *Graph {
	return &Graph{Adj: make(map[string]map[string]bool)}
}

func (g *Graph) AddNode(n string) {
	if g.Adj[n] == nil {
		g.Adj[n] = make(map[string]bool)
	}
}

func (g *Graph) AddUndirectedEdge(a, b string) {
	g.AddNode(a)
	g.AddNode(b)
	g.Adj[a][b] = true
	g.Adj[b][a] = true
}

func (g *Graph) Neighbors(n string) []string {
	m := g.Adj[n]
	if m == nil {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
