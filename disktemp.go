// Copyright © 2019 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func parseDiskTemp() {
	sysBlock := "/sys/block"
	devices, err := ioutil.ReadDir(sysBlock)
	if err != nil {
		return
	}

	for _, dev := range devices {
		fullpath := path.Join(sysBlock, dev.Name())
		link, err := os.Readlink(fullpath)
		if err != nil {
			continue
		}

		if strings.HasPrefix(link, "../devices/virtual/") || slurpFile(path.Join(fullpath, "removable")) == "1" {
			continue
		}

		if strings.HasPrefix(dev.Name(), "nvme") {
			cmd := exec.Command("/usr/sbin/nvme", "smart-log", "/dev/"+dev.Name())
			cmd.Stderr = ioutil.Discard
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				continue
			}
			if err = cmd.Start(); err != nil {
				continue
			}

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				re := regexp.MustCompile(`^temperature\s+: (\d+) C$`)
				if m := re.FindStringSubmatch(scanner.Text()); m != nil {
					temp, err := strconv.ParseFloat(m[1], 64)
					if err != nil {
						continue
					}
					fmt.Printf("%-32s %5.1f\n", "nvmetemp."+dev.Name(), temp)
				}
			}

			cmd.Wait()
		} else {
			var out bytes.Buffer
			cmd := exec.Command("/usr/sbin/hddtemp", "-nw", "/dev/"+dev.Name())
			cmd.Stdout = &out
			cmd.Stderr = ioutil.Discard
			if err = cmd.Run(); err != nil {
				continue
			}

			temp, err := strconv.ParseFloat(strings.TrimRight(out.String(), "\n"), 64)
			if err != nil {
				continue
			}

			fmt.Printf("%-32s %5.1f\n", "hddtemp."+dev.Name(), temp)
		}
	}
}
