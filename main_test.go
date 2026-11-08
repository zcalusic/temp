// Copyright © 2026 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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

// output returns what run prints for the tree at root.
func output(root string) string {
	var out bytes.Buffer
	run(root, &out)

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
			name: "output is sorted as text across devices",
			tree: tree{files: map[string]string{
				"hwmon0/name":        "xe",
				"hwmon0/temp1_input": "44000",
				"hwmon0/temp1_label": "vram_ch_2",
				"hwmon0/temp2_input": "46000",
				"hwmon0/temp2_label": "vram_ch_10",
				"hwmon1/name":        "acpitz",
				"hwmon1/temp1_input": "42000",
			}},
			want: line("acpitz.1", "42.0") + line("xe.vram_ch_10", "46.0") + line("xe.vram_ch_2", "44.0"),
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
			if got := output(tt.tree.write(t)); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestRunMissingRoot(t *testing.T) {
	if got := output(filepath.Join(t.TempDir(), "missing")); got != "" {
		t.Errorf("got %q, want no output", got)
	}
}

// TestRunLineFormat spells out one line, as line() shares the format with run.
func TestRunLineFormat(t *testing.T) {
	root := tree{files: map[string]string{
		"hwmon0/name":        "acpitz",
		"hwmon0/temp1_input": "42000",
	}}.write(t)

	want := "acpitz.1                          42.0\n"
	if got := output(root); got != want {
		t.Errorf("got %q, want %q", got, want)
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
