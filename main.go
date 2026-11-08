// Copyright © 2019 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

// Temp shows system temperatures, as collected from the Linux hwmon subsystem.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var reInput = regexp.MustCompile(`^temp(\d+)_input$`)

// reading is one temperature sensor's current value.
type reading struct {
	name        string
	qualified   string // name with the device in it, for when name is ambiguous
	millidegree int
}

func main() {
	_ = Version

	run("/sys/class/hwmon", os.Stdout)
}

// run prints the temperatures found in the hwmon tree at root, sorted.
func run(root string, stdout io.Writer) {
	readings := collect(root)
	disambiguate(readings)

	lines := make([]string, 0, len(readings))
	for _, r := range readings {
		lines = append(lines, fmt.Sprintf("%-32s %5.1f\n", r.name, float64(r.millidegree)/1000))
	}

	slices.Sort(lines)

	_, _ = fmt.Fprint(stdout, strings.Join(lines, ""))
}

// collect returns the readings of every device in the hwmon tree at root.
func collect(root string) []reading {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	var readings []reading
	for _, entry := range entries {
		readings = append(readings, readDevice(filepath.Join(root, entry.Name()))...)
	}

	return readings
}

// disambiguate switches to the qualified name wherever different devices share
// a name, e.g. identical memory modules. All other names stay as they are.
func disambiguate(readings []reading) {
	first := make(map[string]string, len(readings))
	ambiguous := make(map[string]bool)

	for _, r := range readings {
		qualified, seen := first[r.name]
		if !seen {
			first[r.name] = r.qualified
		} else if qualified != r.qualified {
			ambiguous[r.name] = true
		}
	}

	for i := range readings {
		if ambiguous[readings[i].name] {
			readings[i].name = readings[i].qualified
		}
	}
}

// readDevice returns the readings of the hwmon device in dir. Sensors that
// can't be read, or that read zero (not connected), are left out.
func readDevice(dir string) []reading {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	device := strings.ToLower(slurpFile(filepath.Join(dir, "name")))
	if device == "" {
		return nil
	}

	id := deviceID(dir)

	var readings []reading
	for _, file := range files {
		input := reInput.FindStringSubmatch(file.Name())
		if input == nil {
			continue
		}
		index := input[1]

		label := strings.ReplaceAll(strings.ToLower(slurpFile(filepath.Join(dir, "temp"+index+"_label"))), " ", "_")

		sensor, ok := sensorName(dir, device, index, label)
		if !ok {
			continue
		}

		millidegree, err := strconv.Atoi(slurpFile(filepath.Join(dir, file.Name())))
		if err != nil || millidegree == 0 {
			continue
		}

		readings = append(readings, reading{
			name:        device + "." + sensor,
			qualified:   device + "." + id + "." + sensor,
			millidegree: millidegree,
		})
	}

	return readings
}

// deviceID identifies the hwmon device in dir among devices of the same name:
// the device it belongs to, or the hwmon directory if it has none.
func deviceID(dir string) string {
	target, err := os.Readlink(filepath.Join(dir, "device"))
	if err != nil {
		return filepath.Base(dir)
	}

	return filepath.Base(target)
}

// sensorName builds the part of the displayed name that follows the device
// name, for sensor index of the hwmon device in dir. It returns false if the
// sensor should not be shown.
func sensorName(dir, device, index, label string) (string, bool) {
	sensor := index
	if label != "" {
		sensor = label
	}

	switch device {
	case "drivetemp":
		// Named after the disk: the device has a single meaningful sensor.
		blockdevs, err := filepath.Glob(filepath.Join(dir, "device", "block", "*"))
		if err != nil || len(blockdevs) != 1 || index != "1" {
			return "", false
		}

		return filepath.Base(blockdevs[0]), true
	case "nvme":
		// The device symlink points to the controller, e.g. nvme0.
		controller, err := os.Readlink(filepath.Join(dir, "device"))
		if err != nil {
			return "", false
		}

		return filepath.Base(controller) + "." + sensor, true
	default:
		return sensor, true
	}
}
