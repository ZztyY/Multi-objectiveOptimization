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
			if Servie[k].IsDom == 3 {
				continue
			}
			if Servie[k].IsDom != 2 {
				// 判断dcCor
				for p := 0; p < DcCorNum; p++ {
					if DcCor[p].S1 == k || DcCor[p].S2 == k {
						canList[i] = append(canList[i], k)
						Servie[k].IsDom = 2
						break
					}
				}
			}
			if Servie[k].IsDom != 2 {
				// 判断QoSCor
				for j := 0; j < QoSCorNum; j++ {
					if Cor[j].S1 == k || Cor[j].S2 == k {
						canList[i] = append(canList[i], k)
						Servie[k].IsDom = 2
						break
					}
				}
			}
			if Servie[k].IsDom != 2 { // //表明还没在canList[i]集合中
				if Servie[k].IsDom == 1 {
					continue
				}
				for j := k + 1; j < stSqlSer+SerNumPtask; j++ {
					if Servie[j].IsDom == 1 {
						continue
					}
					if Bm.ParetoDominatesService(Servie[k].Qos, Servie[j].Qos) {
						Servie[j].IsDom = 1
						count++
					} else if Bm.ParetoDominatesService(Servie[j].Qos, Servie[k].Qos) {
						Servie[k].IsDom = 1
						break
					}
				}
				if Servie[k].IsDom == 0 {
					canList[i] = append(canList[i], k)
				}
			}
		}
	}
	return canList
}
