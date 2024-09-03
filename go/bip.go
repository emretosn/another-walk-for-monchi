package main

import (
	"math"
	"math/big"
	"math/rand"
	"sync"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/dbfv"
	"github.com/tuneinsight/lattigo/v4/drlwe"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/utils"
)

type Enrollment_s struct {
	params    bfv.Parameters
	encoder   bfv.Encoder
	encryptor rlwe.Encryptor
	mfip      [][]int64
	x         []float64
	Y         []float64
	borders   []float64
	k         []byte
}

type BIP_s struct {
	params      bfv.Parameters
	evaluator   bfv.Evaluator
	c_selection *rlwe.Ciphertext
	r_in        []int32
}

type Gate_s struct {
	params        bfv.Parameters
	encoder       bfv.Encoder
	encryptor     rlwe.Encryptor
	x             []int64
	col_selection []int64
}

type Party_s struct {
	params      bfv.Parameters
	encoder     bfv.Encoder
	evaluator   bfv.Evaluator
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
	rtgShare    *drlwe.RTGShare
	cksShare    *drlwe.CKSShare
	c_z         *rlwe.Ciphertext
	c1sShare    *rlwe.Plaintext
	r_in        []int32
	k           []byte
}

func getOptimalT(l uint64, v_max uint64, bfv_N float64) uint64 {
	// Minimum t for no overflow in scalar prod. result:
	// 		log2(t) >= 2*log2(v_max) + log2(l)
	bfv_t := uint64(math.Pow(2, math.Log2(float64(v_max))*2+math.Log2(float64(l))))
	bfv_T := new(big.Int).SetUint64(bfv_t)
	// Find closer t that fulfills two conditions:
	//   1) t+1 is divisible by N (required to have SIMD in BFV)
	//   2) t is prime (security loves primes)
	isPrime := false
	for !isPrime {
		bfv_t = uint64((math.Ceil(float64(bfv_T.Uint64())/(2*bfv_N))+1)*(2*bfv_N) + 1) // Find the next t s.t. t%2N==1
		bfv_T.SetUint64(bfv_t)
		isPrime = bfv_T.ProbablyPrime(0) // Check if it is prime. 100% accurate for t<2^64
	}
	return bfv_T.Uint64()
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

func ColRotKeyGen(PPool []*Party_s) (rotKeySet *rlwe.RotationKeySet) {
	rtg := dbfv.NewRTGProtocol(PPool[0].params)
	galEls := PPool[0].params.GaloisElementsForRowInnerSum()

	rotKeySet = rlwe.NewRotationKeySet(PPool[0].params.Parameters, galEls)

	for _, galEl := range galEls {
		rtgShareCombined := rtg.AllocateShare()
		crp := rtg.SampleCRP(PPool[0].corr_prng)

		for _, pi := range PPool {
			pi.rtgShare = rtg.AllocateShare()
			rtg.GenShare(pi.sk, galEl, crp, pi.rtgShare)
		}
		for _, pi := range PPool {
			rtg.AggregateShares(pi.rtgShare, rtgShareCombined, rtgShareCombined)
		}
		rtg.GenRotationKey(rtgShareCombined, crp, rotKeySet.Keys[galEl])
	}
	return
}

func (Party *Party_s) getFinalScoreCT(BIP *BIP_s, permProbeTempMask []int64) *rlwe.Ciphertext {
	ringDim := Party.params.N()
	halfRing := float64(ringDim / 2)

	permProbeTempMaskPT := Party.optimizedPlaintextMul(permProbeTempMask)
	finalScoreCT := BIP.evaluator.MulNew(BIP.c_selection, permProbeTempMaskPT)

	for i := 0; i < int(math.Log2(float64(halfRing))); i++ {
		rotation := int(math.Pow(2, float64(i)))

		rotatedCT := BIP.evaluator.RotateColumnsNew(finalScoreCT, rotation)
		BIP.evaluator.Add(finalScoreCT, rotatedCT, finalScoreCT)
	}
	return finalScoreCT
}

func (Party *Party_s) optimizedPlaintextMul(arr []int64) *bfv.PlaintextMul {
	plainMask := bfv.NewPlaintextMul(Party.params, Party.params.MaxLevel())
	Party.encoder.EncodeMul(arr, plainMask)
	return plainMask
}

func CKSDecrypt(params bfv.Parameters, P []*Party_s, result *rlwe.Ciphertext) *rlwe.Ciphertext {
	cks := dbfv.NewCKSProtocol(params, 3.19) // Collective public-key re-encryption

	for _, pi := range P {
		pi.cksShare = cks.AllocateShare(params.MaxLevel())
	}

	zero := rlwe.NewSecretKey(params.Parameters)
	cksCombined := cks.AllocateShare(params.MaxLevel())

	for _, pi := range P[1:] {
		cks.GenShare(pi.sk, zero, result, pi.cksShare)
	}

	encOut := bfv.NewCiphertext(params, 1, params.MaxLevel())
	for _, pi := range P {
		cks.AggregateShares(pi.cksShare, cksCombined, cksCombined)
	}
	cks.KeySwitch(result, cksCombined, encOut)

	return encOut
}

func quantizeFeatures(borders []float64, unQuantizedFeatures []float64) []int64 {
	numFeat := len(unQuantizedFeatures)
	quantizedFeatures := make([]int64, 0, numFeat)
	lenBorders := len(borders)

	for i := 0; i < numFeat; i++ {
		feature := unQuantizedFeatures[i]
		count := 0

		for count < lenBorders && borders[count] <= feature {
			count++
		}

		quantizedFeatures = append(quantizedFeatures, int64(count))
	}

	return quantizedFeatures
}

func refTemplate(sampleQ []int64, mfbrTab [][]int64) []int64 {
	var refTempPacked []int64
	for _, i := range sampleQ {
		refTemp := mfbrTab[i]
		refTempPacked = append(refTempPacked, refTemp...)
	}
	return refTempPacked
}

func genVectOfInt(begin, length int) []int {
	vect := make([]int, length)
	for i := range vect {
		vect[i] = begin + i
	}
	return vect
}

func vectIntPermutation(seed int64, begin, length int) []int {
	r := rand.New(rand.NewSource(seed))
	seed = r.Int63()

	permuted := genVectOfInt(begin, length)

	rand.Shuffle(len(permuted), func(i, j int) {
		permuted[i], permuted[j] = permuted[j], permuted[i]
	})

	return permuted
}

func addSameValToVector(vect []int, val int) []int {
	for i := range vect {
		vect[i] += val
	}
	return vect
}

func genPermutationsConcat(seed int64, nPerm, lenPerm int) []int {
	permutations := make([]int, 0, nPerm*lenPerm)

	for i := 0; i < nPerm; i++ {
		perm := vectIntPermutation(seed+int64(i), 0, lenPerm)
		perm = addSameValToVector(perm, i*lenPerm)
		permutations = append(permutations, perm...)
	}

	return permutations
}

// UNDERFLOW ISSUE
func divideIntoParts(value int32, d int) []int32 {
	parts := make([]int32, d)

	if d == 1 {
		parts[0] = value
		return parts
	}

    // FIND A WAY TO INCORPORATE THE REMAINING
	r := rand.New(rand.NewSource(42))
	remaining := value
	for i := 0; i < d-1; i++ {
		// Generate a random number between 0 and remaining value
		parts[i] = r.Int31() % (1 << 16)
		remaining -= parts[i]
	}

	parts[d-1] = remaining

	rand.Shuffle(d, func(i, j int) { parts[i], parts[j] = parts[j], parts[i] })

	return parts
}

func (Enrollment *Enrollment_s) encryptPermutedRefTempSingleCT(r_values []int32, refTemp []int64, permutations []int) *rlwe.Ciphertext {
	ringDim := Enrollment.params.N()
	permRefTemp := make([]int64, ringDim)

	j := -1
	for i := range refTemp {
		if i%8 == 0 {
			j++
		}
		permRefTemp[i] = refTemp[permutations[i]] + int64(r_values[j])
	}

	ptxt := bfv.NewPlaintext(Enrollment.params, Enrollment.params.MaxLevel())
	Enrollment.encoder.Encode(permRefTemp, ptxt)
	ctxt := bfv.NewCiphertext(Enrollment.params, 1, Enrollment.params.MaxLevel())
	Enrollment.encryptor.Encrypt(ptxt, ctxt)

	return ctxt
}

func getIndexInVect(permutations []int, value int) int {
	for i, v := range permutations {
		if v == value {
			return i
		}
	}
	return -1
}

func getPermutationsInverse(permutations []int) []int {
	nLen := len(permutations)
	permutationsInverse := make([]int, nLen)

	var wg sync.WaitGroup
	wg.Add(nLen)

	for i := 0; i < nLen; i++ {
		go func(i int) {
			defer wg.Done()
			permutationsInverse[i] = getIndexInVect(permutations, i)
		}(i)
	}

	wg.Wait()

	return permutationsInverse
}

func genPermProbeTemplateFromPermInv(quantizedProb []int64, permutationsInverse []int, lenRow int) []int {
	dim := len(quantizedProb)
	permProbeTemp := make([]int, dim)

	var wg sync.WaitGroup
	wg.Add(dim)

	for i := 0; i < dim; i++ {
		go func(i int) {
			defer wg.Done()
			val := int(quantizedProb[i]) + i*lenRow
			permProbeTemp[i] = permutationsInverse[val]
		}(i)
	}

	wg.Wait()
	return permProbeTemp
}

func getPermutedProbeTempMask(permProbeTemp []int, ringDim int) []int64 {
	nFeat := len(permProbeTemp)
	permProbeTempMask := make([]int64, ringDim)

	var wg sync.WaitGroup
	wg.Add(nFeat)

	for i := 0; i < nFeat; i++ {
		go func(i int) {
			defer wg.Done()
			permProbeTempMask[permProbeTemp[i]] = 1
		}(i)
	}

	wg.Wait()

	return permProbeTempMask
}
