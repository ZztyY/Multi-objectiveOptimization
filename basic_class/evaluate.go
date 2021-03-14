package basic_class

func CountPFRate(method1 []BasicSolution, method2 []BasicSolution) float64 {
	count := 0.0
	for i := 0; i < len(method2); i++ {
		m2 := method2[i]
		for j := 0; j < len(method1); j++ {
			m1 := method1[j]
			if Bm.ParetoDominatesWithConstraints(m1.Objective, m2.Objective) {
				count++
				break
			}
		}
	}
	count = count / float64(len(method2))
	return count
}
