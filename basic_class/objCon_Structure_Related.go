package basic_class

func CalStTime(p []int, service1 []Service) []float64 {
	st := make([]float64, ActNum+1)
	fun := TaskNumPro
	for q := 0; q < ProcessNum; q++ {
		for j := 0; j < fun; j++ {
			if j == 0 {
				st[q*fun+j] = 0
			} else {
				st[q*fun+j] = st[q*fun+j-1] + service1[p[q*fun+j-1]].Qos[0]
			}
		}
	}
	return st
}
