// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"

	"encoding/hex"
	"errors"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/eth/crypto/ecies"
	"github.com/dappledger/AnnChain/genesis/eth/crypto/secp256k1"
	"github.com/dappledger/AnnChain/genesis/eth/crypto/sha3"
	"github.com/dappledger/AnnChain/genesis/eth/rlp"
	"golang.org/x/crypto/ripemd160"
)

const (
	pubkeyCompressed   byte = 0x2 // y_bit + x coord
	pubkeyUncompressed byte = 0x4 // x coord + y coord
)

func isOdd(a *big.Int) bool {
	return a.Bit(0) == 1
}

func Keccak256(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func Keccak256Hash(data ...[]byte) (h ethcmn.Hash) {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Deprecated: For backward compatibility as other packages depend on these
func Sha3(data ...[]byte) []byte          { return Keccak256(data...) }
func Sha3Hash(data ...[]byte) ethcmn.Hash { return Keccak256Hash(data...) }

// Creates an ethereum address given the bytes and the nonce
func CreateAddress(b ethcmn.Address, nonce uint64) ethcmn.Address {
	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
	return ethcmn.BytesToAddress(Keccak256(data)[12:])
}

func Sha256(data []byte) []byte {
	hash := sha256.Sum256(data)

	return hash[:]
}

func Ripemd160(data []byte) []byte {
	ripemd := ripemd160.New()
	ripemd.Write(data)

	return ripemd.Sum(nil)
}

// Ecrecover returns the public key for the private key that was used to
// calculate the signature.
//
// Note: secp256k1 expects the recover id to be either 0, 1. Ethereum
// signatures have a recover id with an offset of 27. Callers must take
// this into account and if "recovering" from an Ethereum signature adjust.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

// New methods using proper ecdsa keys from the stdlib
func ToECDSA(prv []byte) *ecdsa.PrivateKey {
	if len(prv) == 0 {
		return nil
	}

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = secp256k1.S256()
	priv.D = ethcmn.BigD(prv)
	priv.PublicKey.X, priv.PublicKey.Y = secp256k1.S256().ScalarBaseMult(prv)
	return priv
}

func ToECDSAPub(pub []byte) *ecdsa.PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(secp256k1.S256(), pub)
	return &ecdsa.PublicKey{Curve: secp256k1.S256(), X: x, Y: y}
}

func ToECDSAPubCompressed(pub []byte) (*ecdsa.PublicKey, error) {
	var err error
	curve := secp256k1.S256()
	byteLen := (curve.Params().BitSize+7)>>3 + 1
	if len(pub) != byteLen {
		return nil, errors.New("invalid pubkey length")
	}

	pk := &ecdsa.PublicKey{Curve: secp256k1.S256()}

	format := pub[0]
	ybit := (format & 0x1) == 0x1
	format &= ^byte(0x1)

	// format is 0x2 | solution, <X coordinate>
	// solution determines which solution of the curve we use.
	/// y^2 = x^3 + Curve.B
	if format != pubkeyCompressed {
		return nil, fmt.Errorf("invalid magic in compressed "+"pubkey string: %d", pub[0])
	}
	pk.X = new(big.Int).SetBytes(pub[1:33])
	pk.Y, err = decompressPoint(curve, pk.X, ybit)
	if err != nil {
		return nil, err
	}

	if pk.X.Cmp(pk.Curve.Params().P) >= 0 {
		return nil, fmt.Errorf("pubkey X parameter is >= to P")
	}
	if pk.Y.Cmp(pk.Curve.Params().P) >= 0 {
		return nil, fmt.Errorf("pubkey Y parameter is >= to P")
	}
	if !pk.Curve.IsOnCurve(pk.X, pk.Y) {
		return nil, fmt.Errorf("pubkey isn't on secp256k1 curve")
	}
	return pk, nil
}

func FromECDSA(prv *ecdsa.PrivateKey) []byte {
	if prv == nil {
		return nil
	}
	return prv.D.Bytes()
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

func FromECDSAPubCompressed(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	curve := secp256k1.S256()

	byteLen := (curve.Params().BitSize+7)>>3 + 1

	b := make([]byte, 0, byteLen)
	format := pubkeyCompressed
	if isOdd(pub.Y) {
		format |= 0x1
	}
	b = append(b, format)
	return paddedAppend(20, b, pub.X.Bytes())
}

func CompressPubkey(before []byte) []byte {
	curve := secp256k1.S256()
	byteLen := (curve.Params().BitSize+7)>>3 + 1

	if len(before) == 2*byteLen+1 || before[0] != 4 {
		return nil
	}

	return FromECDSAPubCompressed(ToECDSAPub(before))
}

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	if len(b) != 32 {
		return nil, errors.New("invalid length, need 256 bits")
	}
	return ToECDSA(b), nil
}

// LoadECDSA loads a secp256k1 private key from the given file.
// The key data is expected to be hex-encoded.
func LoadECDSA(file string) (*ecdsa.PrivateKey, error) {
	buf := make([]byte, 64)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	if _, err := io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	return ToECDSA(key), nil
}

// SaveECDSA saves a secp256k1 private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func SaveECDSA(file string, key *ecdsa.PrivateKey) error {
	k := hex.EncodeToString(FromECDSA(key))
	return ioutil.WriteFile(file, []byte(k), 0600)
}

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
}

// ValidateSignatureValues verifies whether the signature values are valid with
// the given chain rules. The v value is assumed to be either 0 or 1.
func ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {
	if r.Cmp(ethcmn.Big1) < 0 || s.Cmp(ethcmn.Big1) < 0 {
		return false
	}
	// reject upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	if homestead && s.Cmp(secp256k1.HalfN) > 0 {
		return false
	}
	// Frontier: allow s to be in full N range
	return r.Cmp(secp256k1.N) < 0 && s.Cmp(secp256k1.N) < 0 && (v == 0 || v == 1)
}

func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	s, err := Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.Unmarshal(secp256k1.S256(), s)
	return &ecdsa.PublicKey{Curve: secp256k1.S256(), X: x, Y: y}, nil
}

// Sign calculates an ECDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given hash cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(data []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	if len(data) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(data))
	}

	seckey := ethcmn.LeftPadBytes(prv.D.Bytes(), prv.Params().BitSize/8)
	defer zeroBytes(seckey)
	sig, err = secp256k1.Sign(data, seckey)
	return
}

func Encrypt(pub *ecdsa.PublicKey, message []byte) ([]byte, error) {
	return ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(pub), message, nil, nil)
}

func Decrypt(prv *ecdsa.PrivateKey, ct []byte) ([]byte, error) {
	key := ecies.ImportECDSA(prv)
	return key.Decrypt(rand.Reader, ct, nil, nil)
}

func PubkeyToAddress(p ecdsa.PublicKey) ethcmn.Address {
	pubBytes := FromECDSAPub(&p)
	return ethcmn.BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

// decompressPoint decompresses a point on the given curve given the X point and
// the solution to use.
func decompressPoint(curve elliptic.Curve, x *big.Int, ybit bool) (*big.Int, error) {
	// TODO: This will probably only work for secp256k1 due to
	// optimizations.

	// Y = +-sqrt(x^3 + B)
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Add(x3, curve.Params().B)

	// now calculate sqrt mod p of x2 + B
	// This code used to do a full sqrt based on tonelli/shanks,
	// but this was replaced by the algorithms referenced in
	// https://bitcointalk.org/index.php?topic=162805.msg1712294#msg1712294
	pPlus1Div4 := new(big.Int).Div(new(big.Int).Add(curve.Params().P, big.NewInt(1)), big.NewInt(4))
	y := new(big.Int).Exp(x3, pPlus1Div4, curve.Params().P)

	if ybit != isOdd(y) {
		y.Sub(curve.Params().P, y)
	}
	if ybit != isOdd(y) {
		return nil, fmt.Errorf("ybit doesn't match oddness")
	}
	return y, nil
}
