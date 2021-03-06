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

// 获得各组帕累托前沿rank,每次存储第i层前沿
func (self *BasicMooFunc) PartitionIntoRanksService(inds []Service) int {
	rankNum := 1
	for {
		if len(inds) <= 0 {
			break
		}
		var front []Service
		var nonFront []Service

		front = append(front, inds[0]) // 先把0号元素放进去
		Servie[inds[0].Num].F = float64(rankNum)

		// iterate over all the remaining individuals
		for i := 1; i < len(inds); i++ {
			ind := inds[i]
			noOneWasBatter := true

			// iterate over the entire front
			comfrontNum := 0
			for {
				if comfrontNum >= len(front) {
					break
				}
				frontMember := front[comfrontNum]

				if Bm.ParetoDominatesService(frontMember.Qos, ind.Qos) {
					nonFront = append(nonFront, ind)
					noOneWasBatter = false
					break // ind为非前沿解
				} else if Bm.ParetoDominatesService(ind.Qos, frontMember.Qos) {
					front = append(front[:comfrontNum], front[comfrontNum+1:]...) // todo ?
					nonFront = append(nonFront, frontMember)
				} else {
					comfrontNum++
				}
			}
			if noOneWasBatter {
				front = append(front, ind)
				Servie[ind.Num].F = float64(rankNum)
			}
		}
		// build inds out of remainder
		inds = nonFront
		rankNum++
	}
	return rankNum - 1
}

func (self *BasicMooFunc) ParetoDominatesWithConstraints(tempfit1 []float64, tempfit2 []float64) bool {
	f := true
	sNum := 0
	// 若各目标值相同，则不能算支配
	for i := 0; i < len(tempfit1); i++ {
		if math.Abs(tempfit1[i]-tempfit2[i]) < 0.0000001 {
			sNum++
		}
	}
	if sNum == len(tempfit1) {
		return false
	}

	// 判断是否有约束违反
	if tempfit1[NrObj] == 1 && tempfit2[NrObj] == 1 {
		// 表明没有约束违反，根据目标集合判断支配关系
		// 如果不是相同的情况下，分别判断时间、成本越小越好
		for i := 0; i < NrObj; i++ {
			if Obj[i].ObjType == 0 {
				if tempfit1[i] > tempfit2[i] {
					f = false
					return f
				}
			} else {
				if tempfit1[i] < tempfit2[i] {
					f = false
					return f
				}
			}
		}
	} else if tempfit1[NrObj] == 1 && tempfit2[NrObj] == 0 {
		return true
	} else if tempfit1[NrObj] == 0 && tempfit2[NrObj] == 1 {
		return false
	} else {
		if tempfit1[NrObj] > tempfit2[NrObj] {
			return true
		} else {
			return false
		}
	}
	return f
}
