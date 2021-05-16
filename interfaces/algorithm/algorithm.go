package algorithm

import (
	"Multi-objectiveOptimization/MOEAs"
	"Multi-objectiveOptimization/basic_class"
	"Multi-objectiveOptimization/interfaces/errorcode"
	"Multi-objectiveOptimization/interfaces/response"
	"Multi-objectiveOptimization/util"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"math"
	"os"
	"strconv"
	"strings"
)

func Execute(c *gin.Context) {
	processNum := c.DefaultPostForm("process-num", "")
	serviceNum := c.DefaultPostForm("service-num", "")
	taskNum := c.DefaultPostForm("task-num", "")
	constraintNum := c.DefaultPostForm("constraint-num", "")
	objectNum := c.DefaultPostForm("object-num", "")
	objectFile := c.DefaultPostForm("object-file", "")
	soaFile := c.DefaultPostForm("soa-file", "")
	qosNum := c.DefaultPostForm("qos-num", "")
	if processNum == "" || serviceNum == "" || taskNum == "" || constraintNum == "" || objectNum == "" || qosNum == "" {
		response.SendParamsError(c, nil)
		return
	}

	basic_class.ProcessNum = util.StrToInt(processNum)
	basic_class.TaskNumPro = util.StrToInt(taskNum)
	basic_class.NrObj = util.StrToInt(objectNum) // 目标个数

	objectiveLines := strings.Split(objectFile, "\r\n")
	for i := 0; i < basic_class.NrObj; i++ {
		line := strings.Split(objectiveLines[i], "\t")
		basic_class.Obj[i].Num = i
		basic_class.Obj[i].Name = line[1]
		basic_class.Obj[i].ObjType = util.StrToInt(line[2])
		basic_class.Obj[i].AggreType_inPro = util.StrToInt(line[3])
		basic_class.Obj[i].AggreType_betPro = util.StrToInt(line[4])
	}

	basic_class.SerNumPtask = util.StrToInt(serviceNum)

	soaLines := strings.Split(soaFile, "\n")
	for i := 0; i < basic_class.ProcessNum*basic_class.TaskNumPro*basic_class.SerNumPtask; i++ {
		line := strings.Split(soaLines[i], "\t")
		basic_class.Servie[i].Num = util.StrToInt(line[0])
		basic_class.Servie[i].Qos = make([]float64, basic_class.NrObj)
		basic_class.Servie[i].QosMM = make([]float64, basic_class.NrObj)
		for k := 0; k < basic_class.NrObj; k++ {
			basic_class.Servie[i].Qos[k], _ = strconv.ParseFloat(line[1+k], 64)
		}
		line2, _ := strconv.ParseFloat(line[2], 64)
		basic_class.Servie[i].ChaPenalty = line2 * 0.3
		basic_class.Servie[i].DevPenalty = line2 * 0.1
	}
	for p := 0; p < basic_class.ProcessNum; p++ {
		for k := 0; k < basic_class.TaskNumPro; k++ {
			for j := 0; j < basic_class.SerNumPtask; j++ {
				basic_class.Servie[p*basic_class.TaskNumPro*basic_class.SerNumPtask+k*basic_class.SerNumPtask+j].B = p*basic_class.TaskNumPro + k
				basic_class.Servie[p*basic_class.TaskNumPro*basic_class.SerNumPtask+k*basic_class.SerNumPtask+j].Seq = j
			}
		}
	}

	basic_class.ServiceAttributeDefinition()

	basic_class.ConNum = util.StrToInt(constraintNum) // 约束个数
	for i := 0; i < basic_class.ConNum; i++ {
		basic_class.QosCon[i].Num = i
		basic_class.QosCon[i].QoSType = util.RandomNumber(0, 1)
		basic_class.QosCon[i].ProcessNum = util.RandomNumber(0, basic_class.ProcessNum-1)
		s1 := basic_class.QosCon[i].ProcessNum*basic_class.TaskNumPro + util.RandomNumber(0, basic_class.TaskNumPro-1)
		s2 := basic_class.QosCon[i].ProcessNum*basic_class.TaskNumPro + util.RandomNumber(0, basic_class.TaskNumPro-1)
		basic_class.QosCon[i].StActivity = int(math.Min(float64(s1), float64(s2)))
		basic_class.QosCon[i].EndActivity = int(math.Max(float64(s1), float64(s2)))
		toValue := 0.0
		if basic_class.Obj[basic_class.QosCon[i].QoSType].AggreType_inPro == 1 {
			toValue = 1.0
		}
		for j := basic_class.QosCon[i].StActivity; j < basic_class.QosCon[i].EndActivity; j++ {
			stCan := j * basic_class.SerNumPtask
			sum := 0.0
			for k := 0; k < basic_class.SerNumPtask; k++ {
				sum += basic_class.Servie[stCan+k].Qos[basic_class.QosCon[i].QoSType]
			}
			avg := sum / float64(basic_class.SerNumPtask)
			toValue = basic_class.AggQos(toValue, basic_class.Obj[basic_class.QosCon[i].QoSType].AggreType_inPro, avg)
		}
		basic_class.QosCon[i].ExpectBound = toValue
		if basic_class.Obj[basic_class.QosCon[i].QoSType].ObjType == 0 { // 越小越好
			basic_class.QosCon[i].UlBound = toValue * float64(util.RandomNumber(110, 129)) * 0.01
		} else {
			basic_class.QosCon[i].UlBound = toValue * 0.01 * float64(util.RandomNumber(70, 89))
		}
	}

	basic_class.QoSCorNum = util.StrToInt(qosNum) // QoS关系个数
	for i := 0; i < basic_class.QoSCorNum; i++ {
		basic_class.Cor[i].Num = i
		basic_class.Cor[i].Q = util.RandomNumber(0, basic_class.NrObj-1)
		rt1, rt2 := -1, -1 // 随机选择两个活动
		for {
			rt1 = util.RandomNumber(0, basic_class.ProcessNum-1)*basic_class.TaskNumPro + util.RandomNumber(0, basic_class.TaskNumPro-1)
			rt2 = util.RandomNumber(0, basic_class.ProcessNum-1)*basic_class.TaskNumPro + util.RandomNumber(0, basic_class.TaskNumPro-1)
			if rt1 == rt2 {
				break
			}
		}
		s1 := rt1*basic_class.SerNumPtask + util.RandomNumber(0, basic_class.SerNumPtask-1)
		s2 := rt1*basic_class.SerNumPtask + util.RandomNumber(0, basic_class.SerNumPtask-1)
		basic_class.Cor[i].S1 = s1
		basic_class.Cor[i].S2 = s2
		if basic_class.Obj[basic_class.Cor[i].Q].ObjType == 0 { // 越小越好
			basic_class.Cor[i].Value = basic_class.Servie[s2].Qos[basic_class.Cor[i].Q] * float64(util.RandomNumber(7, 9)) * 0.1 // 在0.7-0.9之间
		} else { // 越大越好
			basic_class.Cor[i].Value = basic_class.Servie[s2].Qos[basic_class.Cor[i].Q] * float64(util.RandomNumber(11, 13)) * 0.1 // 在1.1-1.3之间
		}
	}

	dcCorRate := 10.0
	basic_class.DcCorNum = int(float64(basic_class.ProcessNum*basic_class.TaskNumPro*basic_class.SerNumPtask)*dcCorRate) / 100 // dependence and conflict关系个数
	for i := 0; i < basic_class.DcCorNum; i++ {
		basic_class.DcCor[i].Num = i
		rt1, rt2 := -1, -1 // 随机选择两个活动
		for {
			rt1 = util.RandomNumber(0, basic_class.ProcessNum-1)*basic_class.TaskNumPro + util.RandomNumber(0, basic_class.TaskNumPro-1)
			rt2 = util.RandomNumber(0, basic_class.ProcessNum-1)*basic_class.TaskNumPro + util.RandomNumber(0, basic_class.TaskNumPro-1)
			if rt1 == rt2 {
				break
			}
			s1 := rt1*basic_class.SerNumPtask + util.RandomNumber(0, basic_class.SerNumPtask-1) // 具有DC 关联的服务
			s2 := rt1*basic_class.SerNumPtask + util.RandomNumber(0, basic_class.SerNumPtask-1)
			basic_class.DcCor[i].S1 = s1
			basic_class.DcCor[i].S2 = s2
			basic_class.DcCor[i].DcType = util.RandomNumber(0, 0)
			basic_class.DcCor[i].Flag = true
		}
	}
	basic_class.ActNum = basic_class.ProcessNum * basic_class.TaskNumPro //总活动数目=processNum*taskNumPro
	basic_class.CpTask = 2

	for i := 0; i < 50; i++ {
		basic_class.ExeState.SerNum[i] = -1
	}

	basic_class.RuntimeState = true
	basic_class.IniExePlan.GenBasicSolution(basic_class.ProcessNum, basic_class.TaskNumPro)

	var nsga3 MOEAs.NSGA_3
	nsga3.Init(50, 50, 50, 50, 0.1, 0.1, 0.1, 0.1)
	nsga3.Run()
	//for _, v := range nsga3.MainPop {
	//	logger.LogInfo("Solution", v.Solution)
	//	logger.LogInfo("Objective", v.Objective)
	//	logger.LogInfo("TchVal", v.TchVal)
	//}
	temp, err := os.Create("temp.txt")
	if err != nil {
		response.SendError(c, errorcode.CODE_PARAMS_INVALID, "创建文件失败", nil)
		return
	}
	resu, _ := json.Marshal(nsga3.MainPop)
	_, err = temp.WriteString(string(resu))
	_ = temp.Close()
	c.File("temp.txt")
	_ = os.Remove("temp.txt")
	return
}
