// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cosi

import (
	"crypto/sha512"
	"crypto/subtle"

	//"golang.org/x/crypto/ed25519"
	//"golang.org/x/crypto/ed25519/internal/edwards25519"
	"github.com/bford/golang-x-crypto/ed25519"
	"github.com/bford/golang-x-crypto/ed25519/internal/edwards25519"
)

// Policy represents a fully customizable cosigning policy
// deciding what cosigner sets are and aren't sufficient
// for a collective signature to be considered acceptable to a verifier.
// The Check method may inspect the set of participants that cosigned
// by invoking cosigners.Mask and/or cosigners.MaskBit,
// and may use any other relevant contextual information
// (e.g., how security-critical
// the operation relying on the collective signature is)
// in determining whether the collective signature
// was produced by an acceptable set of cosigners.
type Policy interface {
	Check(cosigners *Cosigners) bool
}

// The default, conservative policy
// just requires all participants to have signed.
type fullPolicy struct{}

func (_ fullPolicy) Check(cosigners *Cosigners) bool {
	return cosigners.CountEnabled() == cosigners.CountTotal()
}

type thresPolicy struct{ t int }

func (p thresPolicy) Check(cosigners *Cosigners) bool {
	return cosigners.CountEnabled() >= p.t
}

// ThresholdPolicy creates a Policy object representing a simple T-of-N policy,
// which deems a collective signature acceptable provided
// that at least the given threshold number of participants cosigned.
func ThresholdPolicy(threshold int) Policy {
	return &thresPolicy{threshold}
}

// Verify determines whether collective signature represented by sig
// is a valid collective signature on the indicated message,
// collectively signed by an acceptable set of cosigners.
// Whether the set of participating cosigners is acceptable
// is determined by the currently-registered Policy.
//
// The default policy conservatively requires all cosigners to participate
// in order for Verify to deem the collective signature acceptable,
// but this policy may be changed by calling SetPolicy
// before invoking Verify.
//
// Verify changes the Cosigners object's participation bitmask
// to the mask carried in the verified signature,
// before invoking the Policy object.
// Thus, a custom Policy can use Cosigners.MaskBit and/or Cosigners.Mask
// to inspect the set of cosigners that actually signed the message,
// and determine the acceptability of the collective signature
// on the basis of the participation mask
// and any other relevant contextual information.
// In addition, after Verify returns,
// the caller can similarly inspect the resulting participation mask
// to determine which specific cosigners did and did not sign.
//
func (cos *Cosigners) Verify(message, sig []byte) bool {

	cosigSize := ed25519.SignatureSize + cos.MaskLen()
	if len(sig) != cosigSize {
		return false
	}

	// Update our mask to reflect which cosigners actually signed
	cos.SetMask(sig[64:])

	// Check that this represents a sufficient set of signers
	if !cos.policy.Check(cos) {
		return false
	}

	return cos.verify(message, sig[:32], sig[:32], sig[32:64], cos.aggr)
}

func (cos *Cosigners) verify(message, aggR, sigR, sigS []byte,
	sigA edwards25519.ExtendedGroupElement) bool {

	if len(sigR) != 32 || len(sigS) != 32 || sigS[31]&224 != 0 {
		return false
	}

	// Compute the digest against aggregate public key and commit
	var aggK [32]byte
	cos.aggr.ToBytes(&aggK)

	h := sha512.New()
	h.Write(aggR)
	h.Write(aggK[:])
	h.Write(message)
	var digest [64]byte
	h.Sum(digest[:0])

	var hReduced [32]byte
	edwards25519.ScReduce(&hReduced, &digest)

	// The public key used for checking is whichever part was signed
	edwards25519.FeNeg(&sigA.X, &sigA.X)
	edwards25519.FeNeg(&sigA.T, &sigA.T)

	var projR edwards25519.ProjectiveGroupElement
	var b [32]byte
	copy(b[:], sigS)
	edwards25519.GeDoubleScalarMultVartime(&projR, &hReduced, &sigA, &b)

	var checkR [32]byte
	projR.ToBytes(&checkR)
	return subtle.ConstantTimeCompare(sigR, checkR[:]) == 1
}

// SetPolicy changes the current Policy object registered
// for this Cosigners object,
// which is used by Verify to determine the acceptability
// of the participant set indicated in a particular collective signature.
// The default policy in any new Cosigners object
// conservatively requires all cosigners to participate in every signature.
// Standard 'T-of-N' threshold-signing policies may be obtained
// by passing a Policy object produced by the ThresholdPolicy function.
// More exotic, arbitrarily customized policies may be used
// by passing any object that implements the Policy interface.
func (cos *Cosigners) SetPolicy(policy Policy) {
	if policy == nil {
		policy = fullPolicy{}
	}
	cos.policy = policy
}

// Verify checks a collective signature on a given message,
// using a given list of public keys and acceptance policy.
//
// If policy is nil, then all cosigners must have participated
// in order for the collective signature to be considered valid.
// Another common policy to specify is ThresholdPolicy(t),
// where t is the threshold-number of cosigners that must participate.
// Obviously t must be greater than 0 to provide any security at all,
// and t cannot be greater than len(publicKeys) or no signature will verify.
//
// This standalone function is the simplest way to verify collective signatures.
// If the caller expects to verify many signatures consecutively
// using the same list of public keys, however,
// it is marginally more efficient to create a Cosigners object
// and use Cosigners.Verify to check each successive signature.
// This efficiency difference is negligible if the number of cosigners is small,
// but may become significant in the case of many cosigners.
func Verify(publicKeys []ed25519.PublicKey, policy Policy,
	message, sig []byte) bool {

	if len(sig) < ed25519.SignatureSize {
		return false
	}
	cos := NewCosigners(publicKeys, sig[64:])
	cos.SetPolicy(policy)
	return cos.Verify(message, sig)
}
