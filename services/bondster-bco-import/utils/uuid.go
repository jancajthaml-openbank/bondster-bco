// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
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
)

const hexLo = "0123456789abcdef"
const hexHi = "0000000000000000111111111111111122222222222222223333333333333333444444444444444455555555555555556666666666666666777777777777777788888888888888889999999999999999aaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbccccccccccccccccddddddddddddddddeeeeeeeeeeeeeeeeffffffffffffffff"

// UUID generates uuid-4
func UUID() string {
	r := make([]byte, 16)
	rand.Read(r)
	return string([]byte{
		hexHi[int(r[0]&0xFF)],
		hexLo[int(r[0]&0x0F)],
		hexHi[int(r[1]&0xFF)],
		hexLo[int(r[1]&0x0F)],
		hexHi[int(r[2]&0xFF)],
		hexLo[int(r[2]&0x0F)],
		hexHi[int(r[3]&0xFF)],
		hexLo[int(r[3]&0x0F)],
		'-',
		hexHi[int(r[4]&0xFF)],
		hexLo[int(r[4]&0x0F)],
		hexHi[int(r[5]&0xFF)],
		hexLo[int(r[5]&0x0F)],
		'-',
		hexHi[int(r[6]&0xFF)],
		hexLo[int(r[6]&0x0F)],
		hexHi[int(r[7]&0xFF)],
		hexLo[int(r[7]&0x0F)],
		'-',
		hexHi[int(r[8]&0xFF)],
		hexLo[int(r[8]&0x0F)],
		hexHi[int(r[9]&0xFF)],
		hexLo[int(r[9]&0x0F)],
		'-',
		hexHi[int(r[10]&0xFF)],
		hexLo[int(r[10]&0x0F)],
		hexHi[int(r[11]&0xFF)],
		hexLo[int(r[11]&0x0F)],
		hexHi[int(r[12]&0xFF)],
		hexLo[int(r[12]&0x0F)],
		hexHi[int(r[13]&0xFF)],
		hexLo[int(r[13]&0x0F)],
		hexHi[int(r[14]&0xFF)],
		hexLo[int(r[14]&0x0F)],
		hexHi[int(r[15]&0xFF)],
		hexLo[int(r[15]&0x0F)],
	})
}
