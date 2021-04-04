package basic_class

import "math"

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

func IGD(solution [][]float64, trueFront [][]float64) float64 {
	sum := 0.0
	for _, v := range trueFront {
		//pos := -1
		dist := 10000000.0
		for i := 0; i < len(solution); i++ {
			d := GetDist(v, solution[i])
			if d < dist {
				dist = d
				//pos = i
			}
		}
		sum += dist
	}
	return sum / float64(len(trueFront))
}

func GetDist(v1 []float64, v2 []float64) float64 {
	dist := 0.0
	for i := 0; i < len(v1); i++ {
		dist += math.Pow(v1[i]-v2[i], 2)
	}
	return math.Sqrt(dist)
}
