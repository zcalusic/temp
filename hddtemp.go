// Copyright © 2019 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func parseHDDTemp() {
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

		var out bytes.Buffer
		cmd := exec.Command("/usr/sbin/hddtemp", "-n", "/dev/"+dev.Name())
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
