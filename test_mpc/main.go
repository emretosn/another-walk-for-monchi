package main

import (
    "fmt"
    "github.com/tuneinsight/lattigo/v4/bfv"
    "github.com/tuneinsight/lattigo/v4/rlwe"

)

func testMul() {
    paramsDef := bfv.PN13QP218
    params, err := bfv.NewParametersFromLiteral(paramsDef)
    if err != nil {
        panic(err)
    }

    encoder := bfv.NewEncoder(params)

    kgen := bfv.NewKeyGenerator(params)
    sk, pk := kgen.GenKeyPair()

    encryptor := bfv.NewEncryptor(params, pk)
    decryptor := bfv.NewDecryptor(params, sk)
    evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{})

    plaintext := bfv.NewPlaintext(params, params.MaxLevel())
    values := []int64{-31113}
    encoder.Encode(values, plaintext)
    ciphertext := encryptor.EncryptNew(plaintext)

    plaintextMultiplier := bfv.NewPlaintext(params, params.MaxLevel())
    multiplierValues := []int64{1}
    encoder.Encode(multiplierValues, plaintextMultiplier)

    evaluator.Mul(ciphertext, plaintextMultiplier, ciphertext)

    decryptedPlaintext := decryptor.DecryptNew(ciphertext)
    decodedValues := encoder.DecodeIntNew(decryptedPlaintext)

    fmt.Println("Original encrypted value:", values)
    fmt.Println("Plaintext multiplier:", multiplierValues)
    fmt.Println("Result after homomorphic multiplication:", decodedValues[:1])
}
