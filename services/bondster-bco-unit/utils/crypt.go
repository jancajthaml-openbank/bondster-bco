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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// fixme not really secure now
var masterKey = []byte("4D92199549E0F2EF009B4160F3582E5528A11A45017F3EF8")

func EncryptString(data string) string {
	return string(Encrypt([]byte(data)))
}

func Encrypt(data []byte) []byte {
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil
	}
	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(data))
	return data
}

func DecryptString(data string) string {
	return string(Decrypt([]byte(data)))
}

func Decrypt(data []byte) []byte {
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil
	}
	if len(data) < aes.BlockSize {
		return nil
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(data, data)
	return data
}
