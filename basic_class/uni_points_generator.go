package basic_class

type UniPointsGenerator struct {
}

func (self *UniPointsGenerator) GetMUniDistributedPoint(m int, h int) [][]float64 {
	buf := make([]int, m)
	for i := 0; i < m; i++ {
		buf[i] = 0
	}
	var result [][]float64 // 包含popsize个点的集合，每个点与理想点形成权重向量
	for {
		arr := make([]float64, m) // 一个目标维的点

		for r := 1; r != m; r++ { // todo
			arr[r-1] = float64(buf[r]-buf[r-1]) / float64(h)
		}
		arr[m-1] = float64(h-buf[m-1]) / float64(h)
		result = append(result, arr)

		var p int
		for p = m - 1; p != 0 && buf[p] == h; p-- { // todo

		}
		if p == 0 {
			break
		}
		buf[p]++
		for p++; p != m; p++ { // todo
			buf[p] = buf[p-1]
		}
	}
	return result
}
