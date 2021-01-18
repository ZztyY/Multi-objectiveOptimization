package basic_class

type MoeaqiBt struct {
}

func (self *MoeaqiBt) ServicePreProcess() [][]int {
	canList := make([][]int, ProcessNum*TaskNumPro)
	for i := 0; i < ProcessNum*TaskNumPro; i++ {
		canList[i] = []int{}
		stSqlSer := i * SerNumPtask
		count := 0

		for k := stSqlSer; k < stSqlSer+SerNumPtask; k++ {
			if servie[k].IsDom == 3 {
				continue
			}
			if servie[k].IsDom != 2 {
				// 判断dcCor
				for p := 0; p < DcCorNum; p++ {
					if dcCor[p].S1 == k || dcCor[p].S2 == k {
						canList[i] = append(canList[i], k)
						servie[k].IsDom = 2
						break
					}
				}
			}
			if servie[k].IsDom != 2 {
				// 判断QoSCor
				for j := 0; j < QoSCorNum; j++ {
					if cor[j].S1 == k || cor[j].S2 == k {
						canList[i] = append(canList[i], k)
						servie[k].IsDom = 2
						break
					}
				}
			}
			if servie[k].IsDom != 2 { // //表明还没在canList[i]集合中
				if servie[k].IsDom == 1 {
					continue
				}
				for j := k + 1; j < stSqlSer+SerNumPtask; j++ {
					if servie[j].IsDom == 1 {
						continue
					}
					if Bm.ParetoDominatesService(servie[k].Qos, servie[j].Qos) {
						servie[j].IsDom = 1
						count++
					} else if Bm.ParetoDominatesService(servie[j].Qos, servie[k].Qos) {
						servie[k].IsDom = 1
						break
					}
				}
				if servie[k].IsDom == 0 {
					canList[i] = append(canList[i], k)
				}
			}
		}
	}
	return canList
}
