package parser

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

var ErrInvalid = errors.New("invalid data format")

type Room struct {
	Name string
	X    int
	Y    int
}

type Input struct {
	Ants   int
	Start  string
	End    string
	Rooms  map[string]Room
	Links  [][2]string
	Raw    []string
}

func Parse(r io.Reader) (Input, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	in := Input{
		Rooms: make(map[string]Room),
		Links: make([][2]string, 0, 256),
		Raw:   make([]string, 0, 256),
	}

	var antsSet bool
	var wantStart bool
	var wantEnd bool

	seenLink := make(map[string]bool)

	for sc.Scan() {
		line := sc.Text()
		in.Raw = append(in.Raw, line)

		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}

		// Comments and directives
		if strings.HasPrefix(trim, "#") {
			if trim == "##start" {
				if in.Start != "" {
					return Input{}, ErrInvalid
				}
				wantStart = true
				wantEnd = false
			} else if trim == "##end" {
				if in.End != "" {
					return Input{}, ErrInvalid
				}
				wantEnd = true
				wantStart = false
			} else {
				// Any unknown command is ignored (as per spec).
			}
			continue
		}

		if !antsSet {
			n, err := strconv.Atoi(trim)
			if err != nil || n <= 0 {
				return Input{}, ErrInvalid
			}
			in.Ants = n
			antsSet = true
			continue
		}

		if isRoomLine(trim) {
			room, err := parseRoom(trim)
			if err != nil {
				return Input{}, ErrInvalid
			}
			if _, ok := in.Rooms[room.Name]; ok {
				return Input{}, ErrInvalid
			}
			in.Rooms[room.Name] = room

			if wantStart {
				in.Start = room.Name
				wantStart = false
			}
			if wantEnd {
				in.End = room.Name
				wantEnd = false
			}
			continue
		}

		if isLinkLine(trim) {
			a, b, err := parseLink(trim)
			if err != nil {
				return Input{}, ErrInvalid
			}
			if a == b {
				continue
			}
			if _, ok := in.Rooms[a]; !ok {
				return Input{}, ErrInvalid
			}
			if _, ok := in.Rooms[b]; !ok {
				return Input{}, ErrInvalid
			}

			key := normalizeUndirectedEdge(a, b)
			if seenLink[key] {
				return Input{}, ErrInvalid
			}
			seenLink[key] = true

			in.Links = append(in.Links, [2]string{a, b})
			continue
		}

		// Unknown non-comment line => invalid
		return Input{}, ErrInvalid
	}

	if err := sc.Err(); err != nil {
		return Input{}, ErrInvalid
	}
	if !antsSet || in.Start == "" || in.End == "" {
		return Input{}, ErrInvalid
	}
	return in, nil
}

func isRoomLine(s string) bool {
	parts := strings.Fields(s)
	return len(parts) == 3 && !strings.Contains(parts[0], "-")
}

func parseRoom(s string) (Room, error) {
	parts := strings.Fields(s)
	if len(parts) != 3 {
		return Room{}, ErrInvalid
	}
	name := parts[0]
	if name == "" || strings.HasPrefix(name, "L") || strings.HasPrefix(name, "#") || strings.ContainsAny(name, " \t") {
		return Room{}, ErrInvalid
	}
	x, err1 := strconv.Atoi(parts[1])
	y, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return Room{}, ErrInvalid
	}
	return Room{Name: name, X: x, Y: y}, nil
}

func isLinkLine(s string) bool {
	return strings.Count(s, "-") == 1 && !strings.ContainsAny(s, " \t")
}

func parseLink(s string) (string, string, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return "", "", ErrInvalid
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalid
	}
	return parts[0], parts[1], nil
}

func normalizeUndirectedEdge(a, b string) string {
	if a < b {
		return a + "-" + b
	}
	return b + "-" + a
}
