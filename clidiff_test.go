package main

import (
	"encoding/json"
	"os"
	"testing"
)

const sampleHelp = `Usage: mytool [command]

A sample CLI tool for deployment

Available Commands:
  deploy      Deploy the application
  status      Show current status
  rollback    Rollback to previous version

Flags:
  -h, --help              help for mytool
  -v, --verbose           Enable verbose output
      --output string     Output format (default "text")
      --timeout int       Timeout in seconds (default "30")
      --dry-run           Dry run mode
`

func TestParseHelpFlags(t *testing.T) {
	snap := ParseHelp("mytool", sampleHelp)
	if snap.Tool != "mytool" {
		t.Fatalf("tool: got %q, want mytool", snap.Tool)
	}
	if len(snap.Flags) < 4 {
		t.Fatalf("flags: got %d, want >=4", len(snap.Flags))
	}
	found := false
	for _, f := range snap.Flags {
		if f.Long == "output" && f.Type == "string" && f.Default == "text" {
			found = true
		}
	}
	if !found {
		t.Fatal("--output string (default text) not parsed")
	}
}

func TestParseHelpSubcommands(t *testing.T) {
	snap := ParseHelp("mytool", sampleHelp)
	if len(snap.Subs) != 3 {
		t.Fatalf("subcommands: got %d, want 3", len(snap.Subs))
	}
	names := map[string]bool{}
	for _, s := range snap.Subs {
		names[s.Name] = true
	}
	for _, want := range []string{"deploy", "status", "rollback"} {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestCompareBreaking(t *testing.T) {
	old := ParseHelp("mytool", sampleHelp)
	modified := "Usage: mytool [command]\n\nAvailable Commands:\n  deploy      Deploy\n  status      Status\n\nFlags:\n  -h, --help              help\n  -v, --verbose           Verbose\n      --output string     Format (default \"json\")\n      --timeout int       Timeout (default \"30\")\n"
	cur := ParseHelp("mytool", modified)
	changes := Compare(old, cur)
	breaking := 0
	for _, c := range changes {
		if c.Severity == "BREAKING" {
			breaking++
		}
	}
	// --dry-run removed, --output default text→json, rollback subcmd removed
	if breaking < 3 {
		t.Errorf("breaking: got %d, want >=3", breaking)
	}
}

func TestCompareMinor(t *testing.T) {
	old := ParseHelp("mytool", sampleHelp)
	expanded := sampleHelp + "      --format string     Output format type\n"
	cur := ParseHelp("mytool", expanded)
	changes := Compare(old, cur)
	minor := 0
	for _, c := range changes {
		if c.Severity == "MINOR" {
			minor++
		}
	}
	if minor != 1 {
		t.Errorf("minor: got %d, want 1", minor)
	}
}

func TestNoChanges(t *testing.T) {
	snap := ParseHelp("mytool", sampleHelp)
	if ch := Compare(snap, snap); len(ch) != 0 {
		t.Errorf("identical snapshots: got %d changes, want 0", len(ch))
	}
}

func TestSnapshotRoundtrip(t *testing.T) {
	snap := ParseHelp("mytool", sampleHelp)
	tmp := t.TempDir() + "/test.snap"
	d, _ := json.MarshalIndent(snap, "", "  ")
	if err := os.WriteFile(tmp, d, 0644); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	var loaded Snapshot
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.Tool != snap.Tool {
		t.Error("tool mismatch")
	}
	if len(loaded.Flags) != len(snap.Flags) {
		t.Errorf("flags: got %d, want %d", len(loaded.Flags), len(snap.Flags))
	}
	if len(loaded.Subs) != len(snap.Subs) {
		t.Errorf("subs: got %d, want %d", len(loaded.Subs), len(snap.Subs))
	}
}
