package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Flag struct {
	Long    string `json:"long"`
	Short   string `json:"short,omitempty"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
}

type Subcommand struct {
	Name string `json:"name"`
}

type Snapshot struct {
	Tool  string       `json:"tool"`
	Flags []Flag       `json:"flags"`
	Subs  []Subcommand `json:"subcommands"`
}

type Change struct {
	Severity string `json:"severity"`
	Category string `json:"category"`
	Detail   string `json:"detail"`
}

func (c Change) String() string {
	return fmt.Sprintf("[%s] %s: %s", c.Severity, c.Category, c.Detail)
}

var flagRe = regexp.MustCompile(
	`(?m)^\s+(-([\w]),\s+)?--(\S+?)(?:\s+(string|int|bool|float|duration))?\s+.*?(?:\(default\s+"?([^")]*?)"?\))?$`,
)

func CaptureHelp(tool string, args ...string) string {
	out, _ := exec.Command(tool, append(args, "--help")...).CombinedOutput()
	return string(out)
}

func ParseHelp(name, text string) Snapshot {
	s := Snapshot{Tool: name}
	for _, m := range flagRe.FindAllStringSubmatch(text, -1) {
		tp := m[4]
		if tp == "" {
			tp = "bool"
		}
		s.Flags = append(s.Flags, Flag{Long: m[3], Short: m[2], Type: tp, Default: m[5]})
	}
	inCmd := false
	subRe := regexp.MustCompile(`^\s{2,4}(\w[\w-]*)\s{2,}`)
	for _, line := range strings.Split(text, "\n") {
		lo := strings.TrimSpace(strings.ToLower(line))
		if strings.Contains(lo, "command") && strings.HasSuffix(lo, ":") {
			inCmd = true
			continue
		}
		if inCmd {
			if strings.TrimSpace(line) == "" {
				inCmd = false
				continue
			}
			if m := subRe.FindStringSubmatch(line); m != nil {
				s.Subs = append(s.Subs, Subcommand{Name: m[1]})
			}
		}
	}
	return s
}

func Compare(old, cur Snapshot) []Change {
	var ch []Change
	nf, of := map[string]Flag{}, map[string]Flag{}
	for _, f := range cur.Flags {
		nf[f.Long] = f
	}
	for _, f := range old.Flags {
		of[f.Long] = f
	}
	for _, f := range old.Flags {
		n, ok := nf[f.Long]
		if !ok {
			ch = append(ch, Change{"BREAKING", "flag-removed", "--" + f.Long})
			continue
		}
		if f.Type != n.Type {
			ch = append(ch, Change{"BREAKING", "type-changed", fmt.Sprintf("--%s: %s→%s", f.Long, f.Type, n.Type)})
		}
		if f.Default != n.Default {
			ch = append(ch, Change{"BREAKING", "default-changed", fmt.Sprintf("--%s: %q→%q", f.Long, f.Default, n.Default)})
		}
	}
	for _, f := range cur.Flags {
		if _, ok := of[f.Long]; !ok {
			ch = append(ch, Change{"MINOR", "flag-added", "--" + f.Long})
		}
	}
	ns, os2 := map[string]bool{}, map[string]bool{}
	for _, s := range cur.Subs {
		ns[s.Name] = true
	}
	for _, s := range old.Subs {
		os2[s.Name] = true
	}
	for _, s := range old.Subs {
		if !ns[s.Name] {
			ch = append(ch, Change{"BREAKING", "subcmd-removed", s.Name})
		}
	}
	for _, s := range cur.Subs {
		if !os2[s.Name] {
			ch = append(ch, Change{"MINOR", "subcmd-added", s.Name})
		}
	}
	return ch
}
