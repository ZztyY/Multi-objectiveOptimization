package basic_class

import "math"

type BasicFunction struct {
}

func AggQos(orgValue float64, aggType int, addValue float64) float64 {
	if aggType == 0 { // 加法聚合
		orgValue += addValue
	} else if aggType == 1 { // 乘法聚合
		orgValue *= addValue
	} else if aggType == 2 { // max聚合
		orgValue = math.Max(orgValue, addValue)
	} else if aggType == 3 { // min聚合
		orgValue = math.Min(orgValue, addValue)
	}
	return orgValue
}

func AggQosEP(EP []int, service1 []Service) []float64 {
	var result []float64
	for k := 0; k < NrObj; k++ {
		var aggQos float64               // 子流程p的最小目标解
		if Obj[k].AggreType_inPro == 0 { // 加法聚合
			aggQos = 0
		} else if Obj[k].AggreType_inPro == 1 { // 乘法聚合
			aggQos = 1
		}
		for i := CpTask; i < ProcessNum*TaskNumPro; i++ {
			aggQos = AggQos(aggQos, Obj[k].AggreType_inPro, service1[EP[i]].Qos[k])
		}
		result = append(result, aggQos)
	}
	return result
}

func TransService(c1 Service) Service {
	var c2 Service
	c2.B = c1.B
	c2.ChaPenalty = c1.ChaPenalty
	c2.DevPenalty = c1.DevPenalty
	c2.F = c1.F
	c2.Num = c1.Num
	c2.Qos = make([]float64, NrObj)
	c2.QosMM = make([]float64, NrObj)
	for i := 0; i < NrObj; i++ {
		c2.Qos[i] = c1.Qos[i]
		c2.QosMM[i] = c1.QosMM[i] // ???  c2.QosMM[i] = c1.Qos[i]
	}
	c2.Seq = c1.Seq
	return c2
}
