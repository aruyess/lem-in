package solver

import (
	"errors"
	"fmt"
	"sort"
)

var ErrStuck = errors.New("stuck")

func Simulate(ants int, paths []Path, start, end string) ([]string, error) {
	if ants <= 0 || len(paths) == 0 {
		return nil, ErrStuck
	}

	assigned := assignAnts(ants, paths)
	nextIndex := make([]int, len(paths))

	occupied := make(map[string]int) // room -> antID, excludes start/end
	finished := 0

	lines := []string{}

	for finished < ants {
		moves := make([]move, 0, ants)

		// Move existing ants forward.
		for _, p := range paths {
			if len(p.Rooms) < 2 {
				continue
			}
			for j := len(p.Rooms) - 2; j >= 1; j-- {
				room := p.Rooms[j]
				antID, ok := occupied[room]
				if !ok {
					continue
				}
				next := p.Rooms[j+1]
				if next != end {
					if _, busy := occupied[next]; busy {
						continue
					}
				}

				delete(occupied, room)
				if next != end {
					occupied[next] = antID
				} else {
					finished++
				}

				moves = append(moves, move{ant: antID, room: next})
			}
		}

		// Launch new ants using assigned per-path queues.
		for pi, p := range paths {
			if len(p.Rooms) < 2 {
				continue
			}
			if nextIndex[pi] >= len(assigned[pi]) {
				continue
			}

			first := p.Rooms[1]
			if first != end {
				if _, busy := occupied[first]; busy {
					continue
				}
			}

			antID := assigned[pi][nextIndex[pi]]
			nextIndex[pi]++

			if first != end {
				occupied[first] = antID
			} else {
				finished++
			}

			moves = append(moves, move{ant: antID, room: first})
		}

		if len(moves) == 0 {
			return nil, ErrStuck
		}

		sort.Slice(moves, func(i, j int) bool { return moves[i].ant < moves[j].ant })

		line := ""
		for i, m := range moves {
			if i > 0 {
				line += " "
			}
			line += fmt.Sprintf("L%d-%s", m.ant, m.room)
		}
		lines = append(lines, line)
	}

	return lines, nil
}

type move struct {
	ant  int
	room string
}

func assignAnts(n int, paths []Path) [][]int {
	assigned := make([][]int, len(paths))
	load := make([]int, len(paths))
	allEqual := true
	baseEdges := paths[0].Edges()
	for i := 1; i < len(paths); i++ {
		if paths[i].Edges() != baseEdges {
			allEqual = false
			break
		}
	}

	if allEqual {
		for antID := 1; antID <= n; antID++ {
			idx := (antID - 1) % len(paths)
			assigned[idx] = append(assigned[idx], antID)
		}
		return assigned
	}

	lastTieIndex := -1
	for antID := 1; antID <= n; antID++ {
		bestScore := paths[0].Edges() + load[0]
		candidates := []int{0}
		for i := 1; i < len(paths); i++ {
			score := paths[i].Edges() + load[i]
			if score < bestScore {
				bestScore = score
				candidates = []int{i}
				continue
			}
			if score == bestScore {
				candidates = append(candidates, i)
			}
		}

		best := candidates[0]
		if len(candidates) > 1 {
			for offset := 1; offset <= len(paths); offset++ {
				idx := (lastTieIndex + offset) % len(paths)
				for _, cand := range candidates {
					if cand == idx {
						best = idx
						lastTieIndex = idx
						offset = len(paths)
						break
					}
				}
			}
		}

		assigned[best] = append(assigned[best], antID)
		load[best]++
	}
	return assigned
}
