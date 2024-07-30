package main

func selectRows(tableMFBR [][]uint64, selections []int) [][]uint64 {
	A := make([][]uint64, len(selections))
	for i, row := range selections {
		A[i] = tableMFBR[row]
	}
	return A
}

func genIndexMaps(indexes []uint64, cols int) [][]uint64 {
	bi := make([][]uint64, len(indexes))
	for i, index := range indexes {
		bi[i] = make([]uint64, cols)
		for j := range bi[i] {
			if j == int(index) {
				bi[i][j] = 1
			} else {
				bi[i][j] = 0
			}
		}
	}
	return bi
}

func compRowsTimesMasks(A [][]uint64, b [][]uint64) [][]uint64 {
	result := make([][]uint64, len(A))
	for i := range A {
		result[i] = make([]uint64, len(A[i]))
		for j := range A[i] {
			result[i][j] = A[i][j] * b[i][j]
		}
	}
	return result
}

func compFlatRowsTimesMasks(A, b []uint64) []uint64 {
    result := make([]uint64, len(A))
    for i := range A {
        result[i] = A[i] * b[i]
    }
    return result
}
