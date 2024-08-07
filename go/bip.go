package main

import (
	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/dbfv"
	"github.com/tuneinsight/lattigo/v4/drlwe"
	"github.com/tuneinsight/lattigo/v4/ring"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/utils"
)

type Enrollment_s struct {
	params    bfv.Parameters
	encoder   bfv.Encoder
	encryptor rlwe.Encryptor
	Y_1       [][]int64
	Y_2       [][]int64
	c_Y_1     [][]*rlwe.Ciphertext
	c_Y_2     [][]*rlwe.Ciphertext
}

type BIP_s struct {
	params    bfv.Parameters
	evaluator bfv.Evaluator
	c_Y       [][]*rlwe.Ciphertext
	c_x       []*rlwe.Ciphertext
	c_z       []*rlwe.Ciphertext
}

type Gate_s struct {
	params    bfv.Parameters
	encoder   bfv.Encoder
	encryptor rlwe.Encryptor
	x         []int64
	c_x       []*rlwe.Ciphertext
	c_z       []*rlwe.Ciphertext
}

type Party_s struct {
	params      bfv.Parameters
	encoder     bfv.Encoder
	sk          *rlwe.SecretKey
	pk          *rlwe.PublicKey
	encryptor   rlwe.Encryptor
	decryptor   rlwe.Decryptor
	corr_prng   *utils.KeyedPRNG
	prng        *utils.KeyedPRNG
	ckgShare    *drlwe.CKGShare
	rlkEphemSk  *rlwe.SecretKey
	rkgShareOne *drlwe.RKGShare
	rkgShareTwo *drlwe.RKGShare
	c_z         []*rlwe.Ciphertext
	c1sShares   []*rlwe.Plaintext
}

func ColPubKeyGen(PPool []*Party_s) *rlwe.PublicKey {
	ckg := dbfv.NewCKGProtocol(PPool[0].params)
	ckgCombined := ckg.AllocateShare()
	for _, pi := range PPool {
		pi.ckgShare = ckg.AllocateShare()
	}
	commonRandomPoly := ckg.SampleCRP(PPool[0].corr_prng)
	for _, pi := range PPool {
		ckg.GenShare(pi.sk, commonRandomPoly, pi.ckgShare)
	}
	for _, pi := range PPool {
		ckg.AggregateShares(pi.ckgShare, ckgCombined, ckgCombined)
	}
	for _, pi := range PPool {
		pi.pk = rlwe.NewPublicKey(pi.params.Parameters)
		ckg.GenPublicKey(ckgCombined, commonRandomPoly, pi.pk)
		pi.encryptor = bfv.NewEncryptor(pi.params, pi.pk)
	}
	return PPool[0].pk
}

func ColRelinKeyGen(PPool []*Party_s) (rlk *rlwe.RelinearizationKey) {
	rkg := dbfv.NewRKGProtocol(PPool[0].params)
	rlk = rlwe.NewRelinearizationKey(PPool[0].params.Parameters, 1)
	_, rkgCombined1, rkgCombined2 := rkg.AllocateShare()
	for _, pi := range PPool {
		pi.rlkEphemSk, pi.rkgShareOne, pi.rkgShareTwo = rkg.AllocateShare()
	}
	crp := rkg.SampleCRP(PPool[0].corr_prng)
	for _, pi := range PPool {
		rkg.GenShareRoundOne(pi.sk, crp, pi.rlkEphemSk, pi.rkgShareOne)
	}
	for _, pi := range PPool {
		rkg.AggregateShares(pi.rkgShareOne, rkgCombined1, rkgCombined1)
	}
	for _, pi := range PPool {
		rkg.GenShareRoundTwo(pi.rlkEphemSk, pi.sk, rkgCombined1, pi.rkgShareTwo)
	}
	for _, pi := range PPool {
		rkg.AggregateShares(pi.rkgShareTwo, rkgCombined2, rkgCombined2)
	}
	rkg.GenRelinearizationKey(rkgCombined1, rkgCombined2, rlk)
	return
}

func (Enrollment *Enrollment_s) EncryptInput(Y [][]int64) (c_Y [][]*rlwe.Ciphertext) {
	K := len(Y)
	l := len(Y[0])

	y_j := make([]int64, K*l)
	ptxt := bfv.NewPlaintext(Enrollment.params, Enrollment.params.MaxLevel())
	c_Y = make([][]*rlwe.Ciphertext, K)

	for k := range c_Y {
		c_Y[k] = make([]*rlwe.Ciphertext, l)
		for i := range c_Y[k] {
			c_Y[k][i] = bfv.NewCiphertext(Enrollment.params, 1, Enrollment.params.MaxLevel())
			for j := range y_j {
				y_j[j] = Y[k][i]
			}
			Enrollment.encoder.Encode(y_j, ptxt)
			Enrollment.encryptor.Encrypt(ptxt, c_Y[k][i])
		}
	}
	return
}

func (BIP *BIP_s) AddCiphertexts(c_Y_1, c_Y_2 []*rlwe.Ciphertext) (c_Y_Sum []*rlwe.Ciphertext) {
	c_Y_Sum = make([]*rlwe.Ciphertext, len(c_Y_1))
	for i := range c_Y_1 {
		if c_Y_2[i] != nil {
			c_Y_Sum[i] = BIP.evaluator.AddNew(c_Y_1[i], c_Y_2[i])
		} else {
			c_Y_Sum[i] = nil
		}
	}
	return
}

func (P *Party_s) C1ShareDecrypt(c_z []*rlwe.Ciphertext) []*rlwe.Plaintext {
	P.c1sShares = make([]*rlwe.Plaintext, len(c_z))
	ringQ := P.params.RingQ()
	sigma := P.params.Sigma()
	for i := range c_z {
		if c_z[i] != nil {
			P.c1sShares[i] = bfv.NewPlaintext(P.params, c_z[i].Level())
			// c1 to NTT domain
			ringQ.NTT(c_z[i].Value[1], P.c1sShares[i].Value)
			// c1 * sk    <NTT domain>
			ringQ.MulCoeffsMontgomery(c_z[i].Value[1], P.sk.Value.Q, P.c1sShares[i].Value)
			// + ei
			GaussianSampler := ring.NewGaussianSampler(P.prng, ringQ, sigma, int(6*sigma))
			ei := GaussianSampler.ReadNew()
			ringQ.NTTLazy(ei, ei)
			// c1 * sk + ei <NTT domain>
			ringQ.Add(P.c1sShares[i].Value, ei, P.c1sShares[i].Value)
		} else {
			P.c1sShares[i] = nil
		}
	}
	return P.c1sShares
}

// TODO: integrate
func (P *Party_s) AggregateAndDecrypt(Pj_c1sShares []*rlwe.Plaintext) (res [][]int64) {
	p_res := make([]*rlwe.Plaintext, len(Pj_c1sShares))
	res = make([][]int64, len(Pj_c1sShares))
	ringQ := P.params.RingQ()
	// Aggregate the shares
	for i, c1sShare := range Pj_c1sShares {
		p_res[i] = bfv.NewPlaintext(P.params, Pj_c1sShares[i].Level())
		// c0 to NTT domain
		ringQ.NTT(P.c_z[i].Value[0], p_res[i].Value)
		// Add Σ c1s_i
		ringQ.Add(P.c1sShares[i].Value, p_res[i].Value, p_res[i].Value) // + c1s_0   <NTT>
		ringQ.Add(c1sShare.Value, p_res[i].Value, p_res[i].Value)       // + c1s_1   <NTT>
		// Mod Q
		ringQ.Reduce(p_res[i].Value, p_res[i].Value)
		// Undo NTT
		ringQ.InvNTT(p_res[i].Value, p_res[i].Value)
		res[i] = make([]int64, P.params.N())
		P.encoder.Decode(p_res[i], res[i]) // TODO: avoid two copies
	}
	return
}
