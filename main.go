package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "snapshot":
		cmdSnapshot()
	case "compare":
		cmdCompare()
	case "guard":
		cmdGuard()
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `clidiff — CLI contract breaking change detector

Commands:
  snapshot <tool> [output.snap]   Capture --help as structured JSON snapshot
  compare  <old.snap> <new.snap>  Diff two snapshots by semver severity
  guard    <baseline.snap> <tool> CI mode: exit 1 on breaking changes`)
}

func cmdSnapshot() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: clidiff snapshot <tool> [output.snap]")
		os.Exit(1)
	}
	tool := os.Args[2]
	out := tool + ".snap"
	if len(os.Args) > 3 {
		out = os.Args[3]
	}
	help := CaptureHelp(tool)
	snap := ParseHelp(tool, help)
	d, _ := json.MarshalIndent(snap, "", "  ")
	if err := os.WriteFile(out, d, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Snapshot saved: %s (%d flags, %d subcommands)\n", out, len(snap.Flags), len(snap.Subs))
}

func cmdCompare() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: clidiff compare <old.snap> <new.snap>")
		os.Exit(1)
	}
	old, err := loadOrDie(os.Args[2])
	new, err2 := loadOrDie(os.Args[3])
	if err != nil || err2 != nil {
		os.Exit(1)
	}
	changes := Compare(old, new)
	if len(changes) == 0 {
		fmt.Println("No changes detected.")
		return
	}
	for _, c := range changes {
		fmt.Println(c)
	}
}

func cmdGuard() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: clidiff guard <baseline.snap> <tool>")
		os.Exit(1)
	}
	base, err := loadOrDie(os.Args[2])
	if err != nil {
		os.Exit(1)
	}
	tool := os.Args[3]
	current := ParseHelp(tool, CaptureHelp(tool))
	changes := Compare(base, current)
	for _, c := range changes {
		fmt.Println(c)
	}
	hasBreaking := false
	for _, c := range changes {
		if c.Severity == "BREAKING" {
			hasBreaking = true
		}
	}
	if hasBreaking {
		fmt.Fprintln(os.Stderr, "\n❌ BREAKING changes detected!")
		os.Exit(1)
	}
	fmt.Println("✅ No breaking changes.")
}

func loadOrDie(path string) (Snapshot, error) {
	d, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return Snapshot{}, err
	}
	var s Snapshot
	if err := json.Unmarshal(d, &s); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return Snapshot{}, err
	}
	return s, nil
}
