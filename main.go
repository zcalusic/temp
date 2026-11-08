// Copyright © 2019 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

// Temp shows system temperatures, as collected from the Linux hwmon subsystem.
package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func main() {
	_ = Version

	hwmonTree := "/sys/class/hwmon"
	dir, err := os.ReadDir(hwmonTree)
	if err != nil {
		return
	}

	reInput := regexp.MustCompile(`^temp(\d+)_input$`)

	var output []string

	for _, entry := range dir {
		subdir, err := os.ReadDir(path.Join(hwmonTree, entry.Name()))
		if err != nil {
			continue
		}

		name := strings.ToLower(slurpFile(path.Join(hwmonTree, entry.Name(), "name")))
		if name == "" {
			continue
		}

		for _, subentry := range subdir {
			input := reInput.FindStringSubmatch(subentry.Name())
			if input == nil {
				continue
			}

			labelFile := path.Join(hwmonTree, entry.Name(), strings.Replace(subentry.Name(), "_input", "_label", 1))
			label := strings.ReplaceAll(strings.ToLower(slurpFile(labelFile)), " ", "_")

			var fullname string
			switch {
			case name == "drivetemp":
				blockdev, err := filepath.Glob(path.Join(hwmonTree, entry.Name(), "device", "block", "*"))
				if err != nil || len(blockdev) != 1 || input[1] != "1" {
					continue
				}
				fullname = name + "." + path.Base(blockdev[0])
			case name == "nvme":
				devpath, err := os.Readlink(path.Join(hwmonTree, entry.Name(), "device"))
				if err != nil {
					continue
				}
				blockdev := path.Base(devpath)
				if label != "" {
					fullname = name + "." + blockdev + "." + label
				} else {
					fullname = name + "." + blockdev + "." + input[1]
				}
			case label != "":
				fullname = name + "." + label
			default:
				fullname = name + "." + input[1]
			}

			tempStr := slurpFile(path.Join(hwmonTree, entry.Name(), subentry.Name()))
			if tempStr == "" || tempStr == "0" {
				continue
			}

			temp, err := strconv.ParseFloat(tempStr, 64)
			if err != nil {
				continue
			}

			output = append(output, fmt.Sprintf("%-32s %5.1f\n", fullname, temp/1000))
		}
	}

	slices.Sort(output)

	for _, out := range output {
		fmt.Print(out)
	}
}
