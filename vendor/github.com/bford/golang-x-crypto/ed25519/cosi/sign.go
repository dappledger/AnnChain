// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cosi

import (
	cryptorand "crypto/rand"
	"crypto/sha512"
	"io"
	"strconv"

	//"golang.org/x/crypto/ed25519"
	//"golang.org/x/crypto/ed25519/internal/edwards25519"
	"github.com/bford/golang-x-crypto/ed25519"
	"github.com/bford/golang-x-crypto/ed25519/internal/edwards25519"
)

// Commitment represents a byte-slice used in the collective signing process,
// which cosigners produce via Commit and send to the leader
// for combination via AggregateCommit.
type Commitment []byte

// SignaturePart represents a byte-slice used in collective signing,
// which cosigners produce via Cosign and send to the leader
// for combination via AggregateSignature.
type SignaturePart []byte

// Secret represents a one-time random secret used
// in collectively signing a single message.
type Secret struct {
	reduced [32]byte
	valid   bool
}

// Commit is invoked by cosigners to produce a one-time commit
// to be used in the collective signing of a single message.
// Producing this commit requires fresh cryptographically random bits,
// which are taken from rand, or from a default source if rand is nil.
//
// On success, Commit returns the commit as a byte-slice
// to be sent to the leader for aggregation via AggregateCommit,
// and a Secret object representing a cryptographic secret
// to be used later in the corresponding call to Cosign.
// Commit fails and returns an error only if rand yields an error.
func Commit(rand io.Reader) (Commitment, *Secret, error) {

	var secretFull [64]byte
	if rand == nil {
		rand = cryptorand.Reader
	}
	_, err := io.ReadFull(rand, secretFull[:])
	if err != nil {
		return nil, nil, err
	}

	var secret Secret
	edwards25519.ScReduce(&secret.reduced, &secretFull)
	secret.valid = true

	// compute R, the individual Schnorr commit to our one-time secret
	var R edwards25519.ExtendedGroupElement
	edwards25519.GeScalarMultBase(&R, &secret.reduced)

	var encodedR [32]byte
	R.ToBytes(&encodedR)
	return encodedR[:], &secret, nil
}

// Cosign signs the message with privateKey and returns a partial signature. It will
// panic if len(privateKey) is not PrivateKeySize.

// Cosign is used by a cosigner to produce its part of a collective signature.
// This operation requires the cosigner's private key,
// the local per-message Secret previously produced
// by the corresponding call to Commit,
// and the aggregate public key and aggregate commit
// that the leader obtained in this signing round
// from AggregatePublicKey and AggregateCommit respectively.
//
// Since it is security-critical that a particular Secret be used only once,
// Cosign invalidates the secret when it is called,
// and panics if called with a previously-used secret.
func Cosign(privateKey ed25519.PrivateKey, secret *Secret, message []byte,
	aggregateK ed25519.PublicKey, aggregateR Commitment) SignaturePart {

	if l := len(privateKey); l != ed25519.PrivateKeySize {
		panic("ed25519: bad private key length: " + strconv.Itoa(l))
	}
	if l := len(aggregateR); l != ed25519.PublicKeySize {
		panic("ed25519: bad aggregateR length: " + strconv.Itoa(l))
	}
	if !secret.valid {
		panic("ed25519: you must use a cosigning Secret only once")
	}

	h := sha512.New()
	h.Write(privateKey[:32])

	var digest1 [64]byte
	var expandedSecretKey [32]byte
	h.Sum(digest1[:0])
	copy(expandedSecretKey[:], digest1[:])
	expandedSecretKey[0] &= 248
	expandedSecretKey[31] &= 63
	expandedSecretKey[31] |= 64

	var hramDigest [64]byte
	h.Reset()
	h.Write(aggregateR)
	h.Write(aggregateK)
	h.Write(message)
	h.Sum(hramDigest[:0])

	var hramDigestReduced [32]byte
	edwards25519.ScReduce(&hramDigestReduced, &hramDigest)

	// Produce our individual contribution to the collective signature
	var s [32]byte
	edwards25519.ScMulAdd(&s, &hramDigestReduced, &expandedSecretKey,
		&secret.reduced)

	// Erase the one-time secret and make darn sure it gets used only once,
	// even if a buggy caller invokes Cosign twice after a single Commit
	secret.reduced = [32]byte{}
	secret.valid = false

	return s[:] // individual partial signature
}

// AggregatePublicKey computes and returns an aggregate public key
// representing the set of cosigners
// currently enabled in the participation bitmask.
// The leader invokes this method during collective signing
// to determine the aggregate public key that needs to be passed
// to the cosigners and supplied to their Cosign operations.
func (cos *Cosigners) AggregatePublicKey() ed25519.PublicKey {
	var keyBytes [32]byte
	cos.aggr.ToBytes(&keyBytes)
	return keyBytes[:]
}

// AggregateCommit is invoked by the leader during collective signing
// to combine all cosigners' individual commits into an aggregate commit,
// which it must pass back to all cosigners for use in their Cosign operations.
// The commits slice must have length equal to the total number of cosigners,
// but AggregateCommit uses only the entries corresponding to cosigners
// that are enabled in the participation mask.
func (cos *Cosigners) AggregateCommit(commits []Commitment) []byte {

	var aggR, indivR edwards25519.ExtendedGroupElement
	var commitBytes [32]byte

	aggR.Zero()
	for i := range cos.keys {
		if cos.MaskBit(i) == Disabled {
			continue
		}

		if l := len(commits[i]); l != ed25519.PublicKeySize {
			return nil
		}
		copy(commitBytes[:], commits[i])
		if !indivR.FromBytes(&commitBytes) {
			return nil
		}
		aggR.Add(&aggR, &indivR)
	}

	var aggRBytes [32]byte
	aggR.ToBytes(&aggRBytes)
	return aggRBytes[:]
}

var scOne = [32]byte{1}

// AggregateSignature is invoked by the leader during collective signing
// to combine all cosigners' individual signature parts
// into a final collective signature.
// The sigParts slice must have length equal to the total number of cosigners,
// but AggregateSignature uses only the entries corresponding to cosigners
// that are enabled in the participation mask,
// which must be identical to the one
// the leader previously used during AggregateCommit.
func (cos *Cosigners) AggregateSignature(aggregateR Commitment, sigParts []SignaturePart) []byte {

	if l := len(aggregateR); l != ed25519.PublicKeySize {
		panic("ed25519: bad aggregateR length: " + strconv.Itoa(l))
	}

	var aggS, indivS [32]byte
	for i := range cos.keys {
		if cos.MaskBit(i) == Disabled {
			continue
		}

		if l := len(sigParts[i]); l != 32 {
			return nil
		}
		copy(indivS[:], sigParts[i])
		edwards25519.ScMulAdd(&aggS, &aggS, &scOne, &indivS)
	}

	mask := cos.Mask()
	cosigSize := ed25519.SignatureSize + len(mask)
	signature := make([]byte, cosigSize)
	copy(signature[:], aggregateR)
	copy(signature[32:64], aggS[:])
	copy(signature[64:], mask)

	return signature
}

// VerifyPart allows the leader to verify an individual cosigner's
// signature part during collective signing.
// This allows the leader to detect if a buggy or malicious cosigner
// produces an invalid signature part
// that might render the final collective signature unusable.
// In such a situation, the leader cannot complete this signing round,
// but can restart the collective signing process (with new commits)
// after excluding the buggy or malicious cosigner.
func (cos *Cosigners) VerifyPart(message, aggR Commitment,
	signer int, indR, indS []byte) bool {

	return cos.verify(message, aggR, indR, indS, cos.keys[signer])
}
