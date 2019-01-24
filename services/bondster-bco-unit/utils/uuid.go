// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"math/rand"
	"time"
)

const hex_low = "0123456789abcdef"
const hex_high = "0000000000000000111111111111111122222222222222223333333333333333444444444444444455555555555555556666666666666666777777777777777788888888888888889999999999999999aaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbccccccccccccccccddddddddddddddddeeeeeeeeeeeeeeeeffffffffffffffff"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func UUID() string {
	var r []byte = make([]byte, 16)
	rand.Read(r)
	var x = [36]byte{
		hex_high[int(r[0]&0xFF)],
		hex_low[int(r[0]&0x0F)],
		hex_high[int(r[1]&0xFF)],
		hex_low[int(r[1]&0x0F)],
		hex_high[int(r[2]&0xFF)],
		hex_low[int(r[2]&0x0F)],
		hex_high[int(r[3]&0xFF)],
		hex_low[int(r[3]&0x0F)],
		'-',
		hex_high[int(r[4]&0xFF)],
		hex_low[int(r[4]&0x0F)],
		hex_high[int(r[5]&0xFF)],
		hex_low[int(r[5]&0x0F)],
		'-',
		hex_high[int(r[6]&0xFF)],
		hex_low[int(r[6]&0x0F)],
		hex_high[int(r[7]&0xFF)],
		hex_low[int(r[7]&0x0F)],
		'-',
		hex_high[int(r[8]&0xFF)],
		hex_low[int(r[8]&0x0F)],
		hex_high[int(r[9]&0xFF)],
		hex_low[int(r[9]&0x0F)],
		'-',
		hex_high[int(r[10]&0xFF)],
		hex_low[int(r[10]&0x0F)],
		hex_high[int(r[11]&0xFF)],
		hex_low[int(r[11]&0x0F)],
		hex_high[int(r[12]&0xFF)],
		hex_low[int(r[12]&0x0F)],
		hex_high[int(r[13]&0xFF)],
		hex_low[int(r[13]&0x0F)],
		hex_high[int(r[14]&0xFF)],
		hex_low[int(r[14]&0x0F)],
		hex_high[int(r[15]&0xFF)],
		hex_low[int(r[15]&0x0F)],
	}

	return string(x[:])
}
