# lem-in (Go)

## What the program does
`lem-in` reads an ant-farm map from a file, finds a plan that moves all ants from `##start` to `##end` in the minimum number of turns, prints the original map, then prints the moves per turn.

Output format (stdout):
1) the exact input map
2) a blank line
3) move lines, one per turn: `Lx-room Lz-room ...`

## Build and run

```bash
go run ./cmd/lem-in <mapfile>
```

Example:
```bash
go run ./cmd/lem-in example00.txt
```

## Notes
- Only standard Go packages are used.
- Unknown commands (lines starting with `#` except `##start` / `##end`) are ignored while parsing, as required.
- The solver uses node-splitting + Edmonds–Karp (max-flow) to get multiple vertex-disjoint paths, then selects the best subset by makespan and simulates the moves.

## Errors
On any invalid input, the program prints:
`ERROR: invalid data format`
