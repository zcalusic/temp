// Copyright © 2019 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func parseHWMon() {
	hwmonTree := "/sys/class/hwmon"
	dir, err := ioutil.ReadDir(hwmonTree)
	if err != nil {
		return
	}

	reInput := regexp.MustCompile(`^temp(\d+)_input$`)

	for _, entry := range dir {
		subdir, err := ioutil.ReadDir(path.Join(hwmonTree, entry.Name()))
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
			label := strings.Replace(strings.ToLower(slurpFile(labelFile)), " ", "_", -1)

			var fullname string
			if name == "drivetemp" {
				blockdev, err := filepath.Glob(path.Join(hwmonTree, entry.Name(), "device", "block", "*"))
				if err != nil || len(blockdev) != 1 || input[1] != "1" {
					continue
				}
				fullname = name + "." + path.Base(blockdev[0])
			} else if name == "nvme" {
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
			} else if label != "" {
				fullname = name + "." + label
			} else {
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

			fmt.Printf("%-32s %5.1f\n", fullname, temp/1000)
		}
	}
}
