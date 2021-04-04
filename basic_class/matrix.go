package basic_class

func IMatrix(matrix [][]float64, dim int) [][]float64 {
	mat := make([][]float64, dim*2)
	for k, _ := range mat {
		mat[k] = make([]float64, dim*2)
	}
	for i := 0; i < dim; i++ {
		for j := 0; j < 2*dim; j++ {
			if j < dim {
				mat[i][j] = matrix[i][j]
			} else {
				if j-dim == i {
					mat[i][j] = 1
				} else {
					mat[i][j] = 0
				}
			}
		}
	}
	for i := 0; i < dim; i++ {
		if mat[i][i] < 1e-6 {
			var j int
			for j = i + 1; j < dim; j++ {
				if mat[j][i] != 0 {
					break
				}
			}
			if j == dim {
				return nil
			}
			for r := i; r < 2*dim; r++ {
				mat[i][r] += mat[j][r]
			}
		}
		ep := mat[i][i]
		for r := i; r < 2*dim; r++ {
			mat[i][r] /= ep
		}

		for j := i + 1; j < dim; j++ {
			e := -1 * (mat[j][i] / mat[i][i])
			for r := i; r < 2*dim; r++ {
				mat[j][r] += e * mat[i][r]
			}
		}
	}

	for i := dim - 1; i >= 0; i-- {
		for j := i - 1; j >= 0; j-- {
			e := -1 * (mat[j][i] / mat[i][i])
			for r := i; r < 2*dim; r++ {
				mat[j][r] += e * mat[i][r]
			}
		}
	}

	result := make([][]float64, dim)
	for k, _ := range result {
		result[k] = make([]float64, dim)
	}
	for i := 0; i < dim; i++ {
		for r := dim; r < 2*dim; r++ {
			result[i][r-dim] = mat[i][r]
		}
	}
	return result
}

func MatrixMultiple(matrix [][]float64, vec []float64) []float64 {
	if matrix == nil {
		return nil
	}
	dim := len(vec)

	v := make([]float64, dim)
	for i := 0; i < dim; i++ {
		v[i] = 0.0
		for j := 0; j < dim; j++ {
			v[i] += matrix[i][j] * vec[j]
		}
	}
	return v
}
