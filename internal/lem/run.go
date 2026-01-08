package lem

import (
	"bytes"
	"os"

	"lem-in/internal/parser"
	"lem-in/internal/solver"
)

func RunFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	in, err := parser.Parse(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}

	paths, err := solver.FindBestPaths(in)
	if err != nil {
		return "", err
	}

	moves, err := solver.Simulate(in.Ants, paths, in.Start, in.End)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	out.Write(raw)

	// Ensure the map ends with a single newline before the blank line.
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		out.WriteByte('\n')
	}
	out.WriteByte('\n')

	for i, line := range moves {
		out.WriteString(line)
		if i != len(moves)-1 {
			out.WriteByte('\n')
		}
	}
	if len(moves) > 0 {
		out.WriteByte('\n')
	}

	return out.String(), nil
}
