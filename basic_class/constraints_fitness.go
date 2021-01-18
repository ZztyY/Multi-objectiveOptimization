package basic_class

import "math"

type ConstraintsFitness struct {
}

func (self *ConstraintsFitness) CalIniConstraint(p []int, corFlag bool) float64 {
	var consNum float64
	service1 := self.UpdateserviceByCor(p, corFlag)
	consNum = 1
	for i := 0; i < ConNum; i++ {
		stask := qosCon[i].StActivity
		etask := qosCon[i].EndActivity
		if etask < CpTask {
			continue
		}
		QoSType := qosCon[i].QoSType
		// pNum := qosCon[i].ProcessNum // 所属流程
		expbound := qosCon[i].ExpectBound
		ubound := qosCon[i].UlBound
		actQos := 0.0
		if Obj[QoSType].ObjType == 1 {
			actQos = 1
		}

		// 计算期望值、底限值目前仅考虑时间、成本约束
		for k := 0; k < NrObj; k++ {
			if QoSType == k {
				for j := stask; j <= etask; j++ {
					actQos = AggQos(actQos, Obj[k].AggreType_inPro, service1[p[j]].Qos[k])
				}
				break
			}
		}

		// 判断约束满足情况
		if Obj[QoSType].ObjType == 0 {
			if actQos > expbound && actQos <= ubound {
				consNum *= (ubound - actQos) / (ubound - expbound)
			} else if actQos > ubound {
				consNum = 0
				break
			}
		} else { //===================Accending 类属性=====================
			if actQos < expbound && actQos >= ubound {
				consNum *= (actQos - ubound) / (expbound - ubound)
			} else if actQos < ubound {
				consNum = 0
				break
			}
		}
	}
	return consNum
}

func (self *ConstraintsFitness) CalFitnessMoo(p []int, corFlag bool, penFlag bool) []float64 {
	var service1 []Service
	service1 = self.UpdateserviceByCor(p, corFlag)
	// 计算多目标值
	objValue := AggQosEP(p, service1)

	// 如果为执行过程中的调整，则考虑与初始计划的偏差
	if runtimeFlag {
		// Step1.1 计算计划中各活动的执行时间   7.31
		st := make([]float64, ProcessNum*TaskNumPro)
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

		// Step1.2 计算与原execution plan 偏移的惩罚量
		penalty := 0.0
		if penFlag {
			for j := CpTask; j < ProcessNum*TaskNumPro; j++ {
				if p[j] != executionPlan.Solution[j] { // 调换服务惩罚
					penalty += servie[executionPlan.Solution[j]].ChaPenalty
				} else if st[j] != executionPlan.STime[j] { //时间推移惩罚
					penalty += math.Abs(st[j]-executionPlan.STime[j]) * servie[executionPlan.Solution[j]].DevPenalty
				}
			}
		}

		// Step1.3 更新QoS 偏移量的惩罚值
		objValue[1] += penalty
	}
	return objValue
}

func (self *ConstraintsFitness) CalTotalConstraint(p []int, conFlag bool, corFlag bool, dcCorFlag bool) float64 {
	conCount := 1.0
	if conFlag {
		conCount = self.CalIniConstraint(p, corFlag)
	}
	if dcCorFlag {
		conCount *= self.CalDCCorConstraint(p)
	}
	return conCount
}

func (self *ConstraintsFitness) CalDCCorConstraint(p []int) float64 {
	consNum := 1.0
	// 计算dependence and conflict 约束
	for i := 0; i < DcCorNum; i++ {
		if dcCor[i].Flag {
			s1 := dcCor[i].S1 // 取出相关服务组合
			s2 := dcCor[i].S2

			if p[servie[s1].B] == s1 { // 如果选中s1
				if dcCor[i].DcType == 1 { // dependence 约束
					if p[servie[s2].B] != s2 {
						consNum = 0
						break
					}
				} else {
					if p[servie[s2].B] == s2 {
						consNum = 0
						break
					}
				}
			}
		}
	}
	return consNum
}

func (self *ConstraintsFitness) CalFitnessMooNormalized(p []int, corFlag bool, penFlag bool) []float64 {
	// 更新参数
	service1 := self.UpdateserviceByCor(p, corFlag)

	// 计算多目标值
	objValue := AggQosEP(p, service1)

	// 如果为执行过程中的调整，则考虑与初始计划的偏差
	if runtimeState {
		// Step1.1 计算计划中各活动的执行时间
		st := CalStTime(p, service1)
		// Step1.2 计算与原execution plan 偏移的惩罚量
		fitMod1 := 0.0
		fitMod2 := 0.0
		for i := 0; i < ActNum; i++ {
			if exeState.SerNum[i] < 0 { // 表明该活动要重新安排
				if p[i] != iniExePlan.Solution[i] {
					fitMod1 = fitMod1 + servie[iniExePlan.Solution[i]].ChaPenalty
				} else if st[i]-iniExePlan.STime[i] > 0.001 || iniExePlan.STime[i]-st[i] > 0.001 {
					fitMod2 = fitMod2 + servie[iniExePlan.Solution[i]].DevPenalty*math.Abs(st[i]-iniExePlan.STime[i])
				}
			}
		}
		// Step1.3 更新QoS 偏移量的惩罚值
		objValue[1] += fitMod1 + fitMod2
	}

	// Step2:计算出目标函数的归一化QoS，先不考虑runtime情况
	for i := 0; i < NrObj; i++ {
		if Obj[i].ObjType == 0 {
			objValue[i] = (qualMinMax[2*i][1] - objValue[i]) / (qualMinMax[2*i][1] - qualMinMax[2*i][0])
		} else {
			objValue[i] = (objValue[i] - qualMinMax[2*i][0]) / (qualMinMax[2*i][1] - qualMinMax[2*i][0])
		}
		objValue[i] = 1 - objValue[i]
	}
	return objValue
}

func (self *ConstraintsFitness) UpdateserviceByCor(p []int, corFlag bool) []Service {
	service1 := make([]Service, ProcessNum*TaskNumPro*SerNumPtask)
	// 已执行活动
	for v := 0; v < CpTask; v++ {
		service1[p[v]] = TransService(servie[p[v]]) // 第v个活动选中的服务
	}

	for v := CpTask; v < ProcessNum*TaskNumPro; v++ {
		service1[p[v]] = TransService(servie[p[v]]) // 第v个活动选中的服务

		if corFlag {
			for j := 0; j < QoSCorNum; j++ {
				if cor[j].S2 == p[v] {
					if p[servie[cor[j].S1].B] == cor[j].S1 {
						qt := cor[j].Q
						if Obj[qt].ObjType == 0 {
							if service1[p[v]].Qos[qt] > cor[j].Value {
								service1[p[v]].Qos[qt] = cor[j].Value
							}
						} else if Obj[qt].ObjType == 1 {
							if service1[p[v]].Qos[qt] < cor[j].Value {
								service1[p[v]].Qos[qt] = cor[j].Value
							}
						}
					}
				}
			}
		}
	}
	return service1
}
