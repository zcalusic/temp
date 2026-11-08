// Copyright © 2026 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tree describes a fixture hwmon tree. All paths are relative to its root.
type tree struct {
	files map[string]string // path to content
	dirs  []string
	links map[string]string // path to symlink target
}

// write creates the tree in a temporary directory and returns its root.
func (tr tree) write(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	mkdir := func(dir string) {
		t.Helper()
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	for _, dir := range tr.dirs {
		mkdir(filepath.Join(root, dir))
	}

	for path, content := range tr.files {
		path = filepath.Join(root, path)
		mkdir(filepath.Dir(path))
		if err := os.WriteFile(path, []byte(content+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for path, target := range tr.links {
		path = filepath.Join(root, path)
		mkdir(filepath.Dir(path))
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

// output returns what run prints for the readable tree at root.
func output(t *testing.T, root string) string {
	t.Helper()

	var out, errOut bytes.Buffer
	if code := run(nil, root, &out, &errOut); code != 0 {
		t.Errorf("exit status %d, want 0", code)
	}

	if errOut.Len() != 0 {
		t.Errorf("unexpected stderr: %q", errOut.String())
	}

	return out.String()
}

// line returns the output line for a sensor.
func line(name, temp string) string {
	return fmt.Sprintf("%-32s %5s\n", name, temp)
}

func TestRun(t *testing.T) {
	tests := []struct {
		name string
		tree tree
		want string
	}{
		{
			name: "no label uses the sensor index",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "acpitz",
				"hwmon0/temp1_input": "42000",
				"hwmon0/temp3_input": "43000",
			}},
			want: line("acpitz.1", "42.0") + line("acpitz.3", "43.0"),
		},
		{
			name: "label is lowercased, spaces become underscores",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "coretemp",
				"hwmon0/temp1_input": "46000",
				"hwmon0/temp1_label": "Package id 0",
			}},
			want: line("coretemp.package_id_0", "46.0"),
		},
		{
			name: "device name is lowercased",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "ACPItz",
				"hwmon0/temp1_input": "42000",
			}},
			want: line("acpitz.1", "42.0"),
		},
		{
			name: "drivetemp is named after the disk, label ignored",
			tree: tree{
				files: map[string]string{
					"hwmon0/name":        "drivetemp",
					"hwmon0/temp1_input": "30000",
					"hwmon0/temp1_label": "ignored",
				},
				dirs: []string{"hwmon0/device/block/sda"},
			},
			want: line("drivetemp.sda", "30.0"),
		},
		{
			name: "drivetemp shows only the first input",
			tree: tree{
				files: map[string]string{
					"hwmon0/name":        "drivetemp",
					"hwmon0/temp1_input": "30000",
					"hwmon0/temp2_input": "31000",
				},
				dirs: []string{"hwmon0/device/block/sda"},
			},
			want: line("drivetemp.sda", "30.0"),
		},
		{
			name: "drivetemp with several disks is skipped",
			tree: tree{
				files: map[string]string{
					"hwmon0/name":        "drivetemp",
					"hwmon0/temp1_input": "30000",
				},
				dirs: []string{"hwmon0/device/block/sda", "hwmon0/device/block/sdb"},
			},
		},
		{
			name: "drivetemp without a disk is skipped",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "drivetemp",
				"hwmon0/temp1_input": "30000",
			}},
		},
		{
			name: "nvme is named after the controller",
			tree: tree{
				files: map[string]string{
					"hwmon0/name":        "nvme",
					"hwmon0/temp1_input": "41000",
					"hwmon0/temp1_label": "Composite",
					"hwmon0/temp2_input": "45000",
				},
				links: map[string]string{"hwmon0/device": "../../nvme0"},
			},
			want: line("nvme.nvme0.2", "45.0") + line("nvme.nvme0.composite", "41.0"),
		},
		{
			name: "nvme without a device link is skipped",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "nvme",
				"hwmon0/temp1_input": "41000",
			}},
		},
		{
			name: "zero, empty, malformed and unreadable readings are dropped",
			tree: tree{
				files: map[string]string{
					"hwmon0/name":        "chip",
					"hwmon0/temp1_input": "0",
					"hwmon0/temp2_input": "",
					"hwmon0/temp3_input": "garbage",
					"hwmon0/temp5_input": "50000",
				},
				dirs: []string{"hwmon0/temp4_input"},
			},
			want: line("chip.5", "50.0"),
		},
		{
			name: "values are formatted to one decimal, negative included",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "chip",
				"hwmon0/temp1_input": "-5500",
				"hwmon0/temp2_input": "1000",
				"hwmon0/temp3_input": "100500",
			}},
			want: line("chip.1", "-5.5") + line("chip.2", "1.0") + line("chip.3", "100.5"),
		},
		{
			name: "files other than temperature inputs are ignored",
			tree: tree{files: map[string]string{
				"hwmon0/name":            "chip",
				"hwmon0/temp1_input":     "50000",
				"hwmon0/temp1_max":       "99000",
				"hwmon0/temp1input":      "98000",
				"hwmon0/temp2_input.old": "97000",
				"hwmon0/xtemp3_input":    "96000",
				"hwmon0/fan1_input":      "1200",
				"hwmon0/in0_input":       "900",
			}},
			want: line("chip.1", "50.0"),
		},
		{
			name: "device without a name is skipped",
			tree: tree{files: map[string]string{"hwmon0/temp1_input": "50000"}},
		},
		{
			name: "entry that is not a directory is skipped",
			tree: tree{files: map[string]string{
				"stray":              "x",
				"hwmon0/name":        "chip",
				"hwmon0/temp1_input": "50000",
			}},
			want: line("chip.1", "50.0"),
		},
		{
			name: "output is in natural order across devices",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "xe",
				"hwmon0/temp1_input": "44000",
				"hwmon0/temp1_label": "vram_ch_2",
				"hwmon0/temp2_input": "46000",
				"hwmon0/temp2_label": "vram_ch_10",
				"hwmon1/name":        "acpitz",
				"hwmon1/temp1_input": "42000",
			}},
			want: line("acpitz.1", "42.0") + line("xe.vram_ch_2", "44.0") + line("xe.vram_ch_10", "46.0"),
		},
		{
			name: "identical devices are told apart by their device",
			tree: tree{
				files: map[string]string{
					"hwmon0/name":        "spd5118",
					"hwmon0/temp1_input": "49000",
					"hwmon1/name":        "spd5118",
					"hwmon1/temp1_input": "48000",
				},
				links: map[string]string{
					"hwmon0/device": "../../../1-0050",
					"hwmon1/device": "../../../1-0051",
				},
			},
			want: line("spd5118.1-0050.1", "49.0") + line("spd5118.1-0051.1", "48.0"),
		},
		{
			name: "only the ambiguous names get the device",
			tree: tree{
				files: map[string]string{
					"hwmon0/name":        "xe",
					"hwmon0/temp1_input": "46000",
					"hwmon0/temp1_label": "pkg",
					"hwmon0/temp2_input": "44000",
					"hwmon0/temp2_label": "vram",
					"hwmon1/name":        "xe",
					"hwmon1/temp1_input": "47000",
					"hwmon1/temp1_label": "pkg",
				},
				links: map[string]string{
					"hwmon0/device": "../../../0000:03:00.0",
					"hwmon1/device": "../../../0000:04:00.0",
				},
			},
			want: line("xe.0000:03:00.0.pkg", "46.0") + line("xe.0000:04:00.0.pkg", "47.0") + line("xe.vram", "44.0"),
		},
		{
			name: "identical devices without a device link use the hwmon directory",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "chip",
				"hwmon0/temp1_input": "49000",
				"hwmon1/name":        "chip",
				"hwmon1/temp1_input": "48000",
			}},
			want: line("chip.hwmon0.1", "49.0") + line("chip.hwmon1.1", "48.0"),
		},
		{
			name: "repeated label within one device stays as it is",
			tree: tree{
				files: map[string]string{
					"hwmon0/name":        "chip",
					"hwmon0/temp1_input": "49000",
					"hwmon0/temp1_label": "same",
					"hwmon0/temp2_input": "48000",
					"hwmon0/temp2_label": "same",
				},
				links: map[string]string{"hwmon0/device": "../../../chip0"},
			},
			want: line("chip.same", "48.0") + line("chip.same", "49.0"),
		},
		{
			name: "empty tree prints nothing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := output(t, tt.tree.write(t)); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestRunUnreadableRoot(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		root string
	}{
		{"missing", filepath.Join(t.TempDir(), "missing")},
		{"not a directory", file},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer

			if code := run(nil, tt.root, &out, &errOut); code != 1 {
				t.Errorf("exit status %d, want 1", code)
			}

			if out.Len() != 0 {
				t.Errorf("unexpected stdout: %q", out.String())
			}

			if got := errOut.String(); !strings.HasPrefix(got, "temp: ") || !strings.Contains(got, tt.root) || !strings.HasSuffix(got, "\n") {
				t.Errorf("stderr %q should be one temp: line naming %s", got, tt.root)
			}
		})
	}
}

// TestRunLineFormat spells out one line, as line() shares the format with run.
func TestRunLineFormat(t *testing.T) {
	root := tree{files: map[string]string{
		"hwmon0/name":        "acpitz",
		"hwmon0/temp1_input": "42000",
	}}.write(t)

	want := "acpitz.1                          42.0\n"
	if got := output(t, root); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestRunArgs uses a missing tree: reading it would exit 1 and print an error,
// so it also proves that handling arguments doesn't touch the tree.
func TestRunArgs(t *testing.T) {
	versionLine := "temp " + Version + "\n"

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"version", []string{"-version"}, 0, versionLine, ""},
		{"version with two dashes", []string{"--version"}, 0, versionLine, ""},
		{"short help", []string{"-h"}, 0, usage, ""},
		{"long help", []string{"-help"}, 0, usage, ""},
		{"unknown flag", []string{"-bogus"}, 2, "", "temp: flag provided but not defined: -bogus\n" + usage},
		{"stray argument", []string{"extra"}, 2, "", "temp: unexpected argument \"extra\"\n" + usage},
		{"stray argument after a flag", []string{"-version", "extra"}, 2, "", "temp: unexpected argument \"extra\"\n" + usage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer

			code := run(tt.args, filepath.Join(t.TempDir(), "missing"), &out, &errOut)
			if code != tt.wantCode {
				t.Errorf("exit status %d, want %d", code, tt.wantCode)
			}

			if got := out.String(); got != tt.wantStdout {
				t.Errorf("stdout %q, want %q", got, tt.wantStdout)
			}

			if got := errOut.String(); got != tt.wantStderr {
				t.Errorf("stderr %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func TestNaturalCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "a", 0},
		{"", "a", -1},
		{"a", "a1", -1},
		{"a", "b", -1},
		{"a2", "a10", -1},
		{"a9b", "a10a", -1},
		{"a1b", "a1c", -1},
		{"a1", "ab", -1},
		{"a2b10", "a2b9", 1},
		{"a01", "a1", 0},
		{"a0", "a00", 0},
		{"a0", "a1", -1},
		{"1-0050", "1-0051", -1},
		{"a99999999999999999999", "a100000000000000000000", -1},
	}

	for _, tt := range tests {
		if got := naturalCompare(tt.a, tt.b); got != tt.want {
			t.Errorf("naturalCompare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}

		if got := naturalCompare(tt.b, tt.a); got != -tt.want {
			t.Errorf("naturalCompare(%q, %q) = %d, want %d", tt.b, tt.a, got, -tt.want)
		}
	}
}

func TestSlurpFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte(" \tvalue \n\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := slurpFile(path); got != "value" {
		t.Errorf("got %q, want %q", got, "value")
	}

	if got := slurpFile(path + ".missing"); got != "" {
		t.Errorf("missing file: got %q, want empty", got)
	}
}
