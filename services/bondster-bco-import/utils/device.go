// Copyright (c) 2016-2023, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"crypto/rand"
	"strconv"
)

const numbers = "0123456789532108"

func checksum(cc string) string {
	var (
		i int
		v uint32 = 0x811c9dc5
		l        = len(cc)
	)

	if l == 0 {
		return "0000000000"
	}

scan:
	v = v ^ uint32(cc[i])&0xFF
	v += (v << 1) + (v << 4) + (v << 7) + (v << 8) + (v << 24)
	i++
	if i != l {
		goto scan
	}

	return strconv.FormatUint(uint64(v>>0), 10)
}

// RandDevice returns random device fingerprint
func RandDevice() string {
	r := make([]byte, 16)
	rand.Read(r)
	device := string([]byte{
		numbers[int(r[0]&0x0F)],
		numbers[int(r[1]&0x0F)],
		numbers[int(r[2]&0x0F)],
		numbers[int(r[3]&0x0F)],
		numbers[int(r[4]&0x0F)],
		numbers[int(r[5]&0x0F)],
		numbers[int(r[6]&0x0F)],
		numbers[int(r[7]&0x0F)],
		numbers[int(r[8]&0x0F)],
		numbers[int(r[9]&0x0F)],
	})
	control := checksum(device)
	return device + "." + control
}
