package basic_class

func Sort(tobeSorted []float64) []int {
	index := make([]int, len(tobeSorted)) // 存储排第i位的数在tobesorted中的索引
	for i := 0; i < len(index); i++ {
		index[i] = i
	}

	for i := 1; i < len(tobeSorted); i++ {
		for j := 0; j < i; j++ {
			if tobeSorted[index[i]] < tobeSorted[index[j]] {
				// insert and break;
				temp := index[i]
				for k := i - 1; k >= j; k-- {
					index[k+1] = index[k]
				}
				index[j] = temp
				break
			}
		}
	}
	return index
}
