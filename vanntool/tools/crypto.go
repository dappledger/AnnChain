/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package tools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"

	"github.com/dappledger/AnnChain/eth/crypto/sha3"
	"golang.org/x/crypto/ed25519"
)

func Encrypt(plantText, key []byte) ([]byte, error) {
	pkey := PaddingLeft(key, '0', 16)
	block, err := aes.NewCipher(pkey) //选择加密算法
	if err != nil {
		return nil, err
	}
	plantText = PKCS7Padding(plantText, block.BlockSize())
	blockModel := cipher.NewCBCEncrypter(block, pkey)
	ciphertext := make([]byte, len(plantText))
	blockModel.CryptBlocks(ciphertext, plantText)
	return ciphertext, nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func DecryptHexText(hexText string, key []byte) ([]byte, error) {
	bytes, err := hex.DecodeString(hexText)
	if err != nil {
		return nil, err
	}
	return Decrypt(bytes, key)
}

func Decrypt(ciphertext, key []byte) ([]byte, error) {
	pkey := PaddingLeft(key, '0', 16)
	block, err := aes.NewCipher(pkey) //选择加密算法
	if err != nil {
		return nil, err
	}
	blockModel := cipher.NewCBCDecrypter(block, pkey)
	plantText := make([]byte, len(ciphertext))
	blockModel.CryptBlocks(plantText, []byte(ciphertext))
	plantText = PKCS7UnPadding(plantText, block.BlockSize())
	return plantText, nil
}

func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

func PaddingLeft(ori []byte, pad byte, length int) []byte {
	if len(ori) >= length {
		return ori[:length]
	}
	pads := bytes.Repeat([]byte{pad}, length-len(ori))
	return append(pads, ori...)
}

func ED25519Pubkey(priv string) (pub string) {
	privBytes, err := hex.DecodeString(priv)
	if err != nil {
		return
	}
	if len(privBytes) < ed25519.PrivateKeySize {
		return
	}
	pubkey := ed25519.PrivateKey(privBytes).Public().(ed25519.PublicKey)
	return hex.EncodeToString(pubkey)
}

func HashKeccak(data []byte) ([]byte, error) {
	keccak := sha3.NewKeccak256()
	if _, err := keccak.Write(data); err != nil {
		return nil, err
	}
	return keccak.Sum(nil), nil
}
