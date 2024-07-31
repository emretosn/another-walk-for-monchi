package main

func selectRows(tableMFBR [][]int64, selections []int) [][]int64 {
	A := make([][]int64, len(selections))
	for i, row := range selections {
		A[i] = tableMFBR[row]
	}
	return A
}

func genIndexMaps(indexes []int64, cols int) [][]int64 {
	bi := make([][]int64, len(indexes))
	for i, index := range indexes {
		bi[i] = make([]int64, cols)
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

func compRowsTimesMasks(A [][]int64, b [][]int64) [][]int64 {
	result := make([][]int64, len(A))
	for i := range A {
		result[i] = make([]int64, len(A[i]))
		for j := range A[i] {
			result[i][j] = A[i][j] * b[i][j]
		}
	}
	return result
}

func compFlatRowsTimesMasks(A, b []int64) []int64 {
    result := make([]int64, len(A))
    for i := range A {
        result[i] = A[i] * b[i]
    }
    return result
}
