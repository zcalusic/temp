# Temp

[![Build Status](https://github.com/zcalusic/temp/actions/workflows/go.yml/badge.svg)](https://github.com/zcalusic/temp/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zcalusic/temp)](https://goreportcard.com/report/github.com/zcalusic/temp)
[![GoDoc](https://godoc.org/github.com/zcalusic/temp?status.svg)](https://godoc.org/github.com/zcalusic/temp)
[![License](https://img.shields.io/badge/license-MIT-a31f34.svg?maxAge=2592000)](https://github.com/zcalusic/temp/blob/master/LICENSE)
[![Powered by](https://img.shields.io/badge/powered_by-Go-5272b4.svg?maxAge=2592000)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux-009bde.svg?maxAge=2592000)](https://www.linuxfoundation.org/)

Temp is a very simple CLI utility that shows system temperatures, as collected from Linux hwmon subsystem.

## Installation

```
go install github.com/zcalusic/temp@latest
```

## Usage

```
temp [-h] [-version]
```

## Sample output

```
acpitz.1                          42.0
coretemp.core_0                   42.0
coretemp.core_1                   46.0
coretemp.package_id_0             46.0
drivetemp.sda                     30.0
iwlwifi.1                         35.0
nvme.nvme0.composite              41.0
pch_wildcat_point.1               38.5
spd5118.1-0050.1                  44.0
spd5118.1-0051.1                  45.5
```

Temperatures are in degrees Celsius. Lines are in natural order, so `core_2` comes before `core_10`.

## Sensor names

Names are lowercase, with spaces replaced by underscores.

| Device       | Name                                                  |
| ------------ | ----------------------------------------------------- |
| SATA disk    | `drivetemp.<disk>`                                    |
| NVMe         | `nvme.<controller>.<label>`, or `<index>` if no label |
| Anything else | `<chip>.<label>`, or `<chip>.<index>` if no label    |

If different devices would end up with the same name, for example identical memory modules, those names also get the device they belong to: `spd5118.1-0050.1`.

Sensors that can't be read, or read exactly 0 (usually not connected), are left out.

SATA disks need the kernel's `drivetemp` module to be loaded (`modprobe drivetemp`).

## Exit status

| Status | Meaning                                                   |
| ------ | --------------------------------------------------------- |
| 0      | Success, including on systems without any sensors         |
| 1      | The hwmon tree (`/sys/class/hwmon`) couldn't be read      |
| 2      | Invalid arguments                                         |

## Contributors

Contributors are welcome, just open a new issue / pull request.

## License

```
The MIT License (MIT)

Copyright © 2019 Zlatko Čalušić

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
