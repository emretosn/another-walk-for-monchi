package main

import (
	"fmt"
	"math/rand"

	"github.com/tuneinsight/lattigo/v4/bfv"
	//"github.com/tuneinsight/lattigo/v4/rlwe"
)

func main() {
    // Two table selection and mask unencrypted
	tabShare1Path := "./lookupTables/Rand/Rand_nB_3_dimF_512.csv"
	tabShare2Path := "./lookupTables/MFIP-Rand/MFIPSubRand_nB_3_dimF_512.csv"

	tabShare1, err := readCSVToArray(tabShare1Path, "[][]int64")
	tabShare2, err := readCSVToArray(tabShare2Path, "[][]int64")
	if err != nil {
		fmt.Println("Error reading one of the csv files:", err)
	}
    fmt.Println("tabShare1")
    printMatrix(tabShare1.([][]int64))
    fmt.Println("tabShare2")
    printMatrix(tabShare2.([][]int64))

    size := len(tabShare1.([][]int64))

    selections := []int{rand.Intn(size), rand.Intn(size), rand.Intn(size)}
    fmt.Printf("selections: %v\n", selections)
    tabShare1selections := selectRows(tabShare1.([][]int64), selections)
    tabShare2selections := selectRows(tabShare2.([][]int64), selections)

    fmt.Println("selection1")
    printMatrix(tabShare1selections)
    fmt.Println("selection2")
    printMatrix(tabShare2selections)

    b := genRandInexes(len(selections), size)
	fmt.Println("Selected Columns", b)

    bmap := genIndexMaps(b, size)

    tabShare1selectionsFlat := flatten(tabShare1selections)
    tabShare2selectionsFlat := flatten(tabShare2selections)
    bmapFlat := flatten(bmap)
    fmt.Println("Flat mask:", bmapFlat)

    result1 := compFlatRowsTimesMasks(tabShare1selectionsFlat, bmapFlat)
    result2 := compFlatRowsTimesMasks(tabShare2selectionsFlat, bmapFlat)

    fmt.Println("result1")
    fmt.Println(result1)
    fmt.Println("result2")
    fmt.Println(result2)

    // Two table selection and mask encrypted
    // TODO: Encrypt the tables not the rows
    fmt.Println("----------ENCRYPTION----------")
    paramDef := bfv.PN13QP218
    paramDef.T = 0x3ee0001
    params, err := bfv.NewParametersFromLiteral(paramDef)
    if err != nil {
        panic(err)
    }

    //tabShare1
    kgen1 := bfv.NewKeyGenerator(params)
    sk1, pk1 := kgen1.GenKeyPair()
    encryptor1pk := bfv.NewEncryptor(params, pk1)
    decryptor1 := bfv.NewDecryptor(params, sk1)
    //tabShare2
    kgen2 := bfv.NewKeyGenerator(params)
    sk2, pk2 := kgen2.GenKeyPair()
    encryptor2pk := bfv.NewEncryptor(params, pk2)
    decryptor2 := bfv.NewDecryptor(params, sk2)

    encoder := bfv.NewEncoder(params)

    //evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{})

    fmt.Println("Flat table1:", result1)
    fmt.Println("Flat table2:", result2)

    // What is the second arg in EncodeNew
    //tabShare1
    plaintext1Poly := encoder.EncodeNew(result1, 0)
    ciphertxt1 := encryptor1pk.EncryptNew(plaintext1Poly)
    //tabShare2
    plaintext2Poly := encoder.EncodeNew(result2, 0)
    ciphertxt2 := encryptor2pk.EncryptNew(plaintext2Poly)

    fmt.Println("Ciphertext1:", ciphertxt1)
    fmt.Println("Ciphertext2:", ciphertxt2)

    fmt.Println(ciphertxt1.Value)

    // Add these two somehow, needs another argument
    //evaluator.Add(ciphertxt1, ciphertxt2)

    decrypted1Poly := decryptor1.DecryptNew(ciphertxt1)
    decrypted2Poly := decryptor2.DecryptNew(ciphertxt2)

    decrypted1Plaintext:= encoder.DecodeIntNew(decrypted1Poly)
    decrypted2Plaintext:= encoder.DecodeIntNew(decrypted2Poly)
    // TODO: Need to adjust the N in parameters for propper size
    // Slicing the decrypted plaintext for now.
    fmt.Println("Decrypted1:", decrypted1Plaintext[:len(result1)])
    fmt.Println("Decrypted2:", decrypted2Plaintext[:len(result2)])
}
