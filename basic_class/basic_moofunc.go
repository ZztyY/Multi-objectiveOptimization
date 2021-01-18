package basic_class

import (
	"math"
)

type BasicMooFunc struct {
}

var Bm = BasicMooFunc{}

func (self *BasicMooFunc) ParetoDominatesService(tempFit1 []float64, tempFit2 []float64) bool {
	f := true
	sNum := 0
	// 若各目标值相同，则不能算支配
	for i := 0; i < NrObj; i++ {
		if tempFit1[i]-tempFit2[i] < 0.0000001 && tempFit1[i]-tempFit2[i] > 0 {
			sNum++
		}
	}
	if sNum == NrObj {
		return false
	}

	// 如果不是相同的情况下，分别判断时间、成本越小越好
	for i := 0; i < NrObj; i++ {
		if Obj[i].ObjType == 0 {
			if tempFit1[i] > tempFit2[i] {
				f = false
				return f
			}
		} else {
			if tempFit1[i] < tempFit2[i] {
				f = false
				return f
			}
		}
	}
	return f
}

func (self *BasicMooFunc) ParetoDominatesMin(tempFit1 []float64, tempFit2 []float64) bool {
	f := true
	sNum := 0
	// 若各目标值相同，则不能算支配
	for i := 0; i < NrObj; i++ {
		if math.Abs(tempFit1[i]-tempFit2[i]) < 0.0000001 {
			sNum++
		}
	}
	if sNum == NrObj {
		return false
	}

	// 如果不是相同的情况下，则越小越好
	for i := 0; i < NrObj; i++ {
		if tempFit1[i] > tempFit2[i] {
			f = false
			return f
		}
	}
	return f
}
