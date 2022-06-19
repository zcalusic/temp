# Temp

[![Build Status](https://travis-ci.org/zcalusic/temp.svg?branch=master)](https://travis-ci.org/zcalusic/temp)
[![Go Report Card](https://goreportcard.com/badge/github.com/zcalusic/temp)](https://goreportcard.com/report/github.com/zcalusic/temp)
[![GoDoc](https://godoc.org/github.com/zcalusic/temp?status.svg)](https://godoc.org/github.com/zcalusic/temp)
[![License](https://img.shields.io/badge/license-MIT-a31f34.svg?maxAge=2592000)](https://github.com/zcalusic/temp/blob/master/LICENSE)
[![Powered by](https://img.shields.io/badge/powered_by-Go-5272b4.svg?maxAge=2592000)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux-009bde.svg?maxAge=2592000)](https://www.linuxfoundation.org/)

Temp is a very simple CLI utility that shows system temperatures, as collected from Linux hwmon subsystem.

## Installation

Just use go get.

```
go get github.com/zcalusic/temp
```

## Sample output

```
acpitz.1                          42.0
pch_wildcat_point.1               38.5
iwlwifi.1                         35.0
coretemp.package_id_0             46.0
coretemp.core_0                   42.0
coretemp.core_1                   46.0
nvme.nvme0.composite              41.0
drivetemp.sda                     30.0
```

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
