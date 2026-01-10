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
	for i, line := range moves {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(line)
	}

	return out.String(), nil
}
