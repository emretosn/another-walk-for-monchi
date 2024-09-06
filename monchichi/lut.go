package main

import (
	"math/rand"
	"sync"
)

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
func getIndexInVect(permutations []int, value int) int {
	for i, v := range permutations {
		if v == value {
			return i
		}
	}
	return -1
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
func genRefTempFromPerm(refTemp []int64, permutations []int, r_values []int32) []int64 {
	dim := len(refTemp)
	permRefTemp := make([]int64, dim)

    j := -1
	for i := 0; i < dim; i++ {
        if i%8 == 0 {
            j++
        }
		permRefTemp[i] = refTemp[permutations[i]] + int64(r_values[j])
	}
	return permRefTemp
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

// create a function that take a vector and create 2 additive shares of it
func createAdditiveShares(vector []int64) ([]int64, []int64) {
	n := len(vector)
	share1 := make([]int64, n)
	share2 := make([]int64, n)
    j := -1
	for i := 0; i < n; i++ {
		if i%8 == 0 {
			j++
		}
		// generate a random number in the range of int32 and cast it to int64
		share1[i] = int64(rand.Int31())
		share2[i] = vector[i] - share1[i]
	}
	// assert that the sum of the shares is equal to the original vector

	return share1, share2
}

// create a lookup function that takes a flatten table and a vector of row indicies and return the corresponding rows
func lookupTable(table []int64, indices []int) int64 {
	n := len(indices)
	result := int64(0)
	for i := 0; i < n; i++ {
		result += table[indices[i]]
	}
	return result
}

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


