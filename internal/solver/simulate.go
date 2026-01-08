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

	// Give each path a contiguous block of ant IDs (matches typical lem-in expected output).
	nextID := make([]int, len(paths))
	remain := make([]int, len(paths))
	cur := 1
	for i := range paths {
		nextID[i] = cur
		remain[i] = assigned[i]
		cur += assigned[i]
	}

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

		// Launch new ants (use per-path ant ID blocks, not global).
		for pi, p := range paths {
			if remain[pi] <= 0 {
				continue
			}
			if len(p.Rooms) < 2 {
				continue
			}

			first := p.Rooms[1]
			if first != end {
				if _, busy := occupied[first]; busy {
					continue
				}
			}

			antID := nextID[pi]
			nextID[pi]++
			remain[pi]--

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

func assignAnts(n int, paths []Path) []int {
	load := make([]int, len(paths))
	for a := 0; a < n; a++ {
		best := 0
		bestScore := paths[0].Edges() + load[0]
		for i := 1; i < len(paths); i++ {
			score := paths[i].Edges() + load[i]
			if score < bestScore {
				bestScore = score
				best = i
			}
		}
		load[best]++
	}
	return load
}
