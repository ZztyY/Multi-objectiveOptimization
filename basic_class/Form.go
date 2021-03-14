package basic_class

import (
	"Multi-objectiveOptimization/util"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
)

type Service struct {
	Num        int       // 服务编号
	B          int       // 服务支持哪一个任务
	Seq        int       // 该任务的第seq个候选
	Qos        []float64 // 服务QoS属性向量
	DevPenalty float64   // 时间偏移惩罚
	ChaPenalty float64   // 调换服务惩罚
	QosMM      []float64 // 对于变迁来说存储最大、最小值，对于普通候选服务，存储标准值与比例值
	F          float64   // 定义自身质量
	IsDom      float64   // 是否被支配，0为不被支配，1为被支配，2为存在关联关系，3为不可用
}

type DcCorrelation struct {
	Num    int
	S1     int
	S2     int
	DcType int
	Flag   bool
}

type QoSCorrelation struct {
	Num   int // 顺序号
	Q     int // QoS属性
	S1    int
	S2    int
	Value float64 // 值
}

type Object struct {
	Num              int
	Name             string
	ObjType          int // 0为decending越小越好，1为accending越大越好
	AggreType_inPro  int // 流程内聚合 0 加法；1乘法聚合
	AggreType_betPro int // 流程间聚合 0 加法；1乘法聚合；2max聚合；3min聚合
}

type QoSConstraint struct {
	Num         int
	QoSType     int
	ProcessNum  int
	StActivity  int
	EndActivity int
	ExpectBound float64
	UlBound     float64 // 上限
}

// ===========Form1===========
var ProcessNum int  // 流程数量（默认为1）
var TaskNumPro int  // 每个流程中的任务数量(由于只考虑一个流程，所以就是任务总数量)
var SerNumPtask int // 每个任务的备选服务数量
var DcCorNum int    // dependence and conflict关系个数
var QoSCorNum int   // QoS关系个数
var NrObj int       // 目标个数
var ActNum int      //总活动数目=processNum*taskNumPro
var ConNum int      // 约束个数
var CpTask int
var dcCor = make([]DcCorrelation, 1500)
var cor = make([]QoSCorrelation, 1500)
var Obj = make([]Object, 10)
var servie = make([]Service, 10000)
var qosCon = make([]QoSConstraint, 1500)
var qualMinMax = [10][2]float64{}

func init() {
	ProcessNum = 1
	TaskNumPro = 50
	NrObj = 4 // 目标个数
	objectiveFile, _ := ioutil.ReadFile("./data/Objective.txt")
	objectiveFileString := string(objectiveFile)
	objectiveLines := strings.Split(objectiveFileString, "\n")
	for i := 0; i < NrObj; i++ {
		line := strings.Split(objectiveLines[i], "\t")
		Obj[i].Num = i
		Obj[i].Name = line[1]
		Obj[i].ObjType = util.StrToInt(line[2])
		Obj[i].AggreType_inPro = util.StrToInt(line[3])
		Obj[i].AggreType_betPro = util.StrToInt(line[4])
	}
	Obj[0].AggreType_betPro = 2
	Obj[1].AggreType_betPro = 0
	Obj[2].AggreType_betPro = 3
	Obj[3].AggreType_betPro = 3

	SerNumPtask = 20
	soaFile, _ := ioutil.ReadFile("./data/SOA.txt")
	soaString := string(soaFile)
	soaLines := strings.Split(soaString, "\n")
	for i := 0; i < ProcessNum*TaskNumPro*SerNumPtask; i++ {
		line := strings.Split(soaLines[i], "\t")
		servie[i].Num = util.StrToInt(line[0])
		servie[i].Qos = make([]float64, NrObj)
		servie[i].QosMM = make([]float64, NrObj)
		for k := 0; k < NrObj; k++ {
			servie[i].Qos[k], _ = strconv.ParseFloat(line[1+k], 64)
		}
		line2, _ := strconv.ParseFloat(line[2], 64)
		servie[i].ChaPenalty = line2 * 0.3
		servie[i].DevPenalty = line2 * 0.1
	}
	for p := 0; p < ProcessNum; p++ {
		for k := 0; k < TaskNumPro; k++ {
			for j := 0; j < SerNumPtask; j++ {
				servie[p*TaskNumPro*SerNumPtask+k*SerNumPtask+j].B = p*TaskNumPro + k
				servie[p*TaskNumPro*SerNumPtask+k*SerNumPtask+j].Seq = j
			}
		}
	}

	serviceAttributeDefinition()

	ConNum = 20 // 约束个数
	for i := 0; i < ConNum; i++ {
		qosCon[i].Num = i
		qosCon[i].QoSType = util.RandomNumber(0, 1)
		qosCon[i].ProcessNum = util.RandomNumber(0, ProcessNum-1)
		s1 := qosCon[i].ProcessNum*TaskNumPro + util.RandomNumber(0, TaskNumPro-1)
		s2 := qosCon[i].ProcessNum*TaskNumPro + util.RandomNumber(0, TaskNumPro-1)
		qosCon[i].StActivity = int(math.Min(float64(s1), float64(s2)))
		qosCon[i].EndActivity = int(math.Max(float64(s1), float64(s2)))
		toValue := 0.0
		if Obj[qosCon[i].QoSType].AggreType_inPro == 1 {
			toValue = 1.0
		}
		for j := qosCon[i].StActivity; j < qosCon[i].EndActivity; j++ {
			stCan := j * SerNumPtask
			sum := 0.0
			for k := 0; k < SerNumPtask; k++ {
				sum += servie[stCan+k].Qos[qosCon[i].QoSType]
			}
			avg := sum / float64(SerNumPtask)
			toValue = AggQos(toValue, Obj[qosCon[i].QoSType].AggreType_inPro, avg)
		}
		qosCon[i].ExpectBound = toValue
		if Obj[qosCon[i].QoSType].ObjType == 0 { // 越小越好
			qosCon[i].UlBound = toValue * float64(util.RandomNumber(110, 129)) * 0.01
		} else {
			qosCon[i].UlBound = toValue * 0.01 * float64(util.RandomNumber(70, 89))
		}
	}

	QoSCorNum = 20 // QoS关系个数
	for i := 0; i < QoSCorNum; i++ {
		cor[i].Num = i
		cor[i].Q = util.RandomNumber(0, NrObj-1)
		rt1, rt2 := -1, -1 // 随机选择两个活动
		for {
			rt1 = util.RandomNumber(0, ProcessNum-1)*TaskNumPro + util.RandomNumber(0, TaskNumPro-1)
			rt2 = util.RandomNumber(0, ProcessNum-1)*TaskNumPro + util.RandomNumber(0, TaskNumPro-1)
			if rt1 == rt2 {
				break
			}
		}
		s1 := rt1*SerNumPtask + util.RandomNumber(0, SerNumPtask-1)
		s2 := rt1*SerNumPtask + util.RandomNumber(0, SerNumPtask-1)
		cor[i].S1 = s1
		cor[i].S2 = s2
		if Obj[cor[i].Q].ObjType == 0 { // 越小越好
			cor[i].Value = servie[s2].Qos[cor[i].Q] * float64(util.RandomNumber(7, 9)) * 0.1 // 在0.7-0.9之间
		} else { // 越大越好
			cor[i].Value = servie[s2].Qos[cor[i].Q] * float64(util.RandomNumber(11, 13)) * 0.1 // 在1.1-1.3之间
		}
	}

	dcCorRate := 10.0
	DcCorNum = int(float64(ProcessNum*TaskNumPro*SerNumPtask)*dcCorRate) / 100 // dependence and conflict关系个数
	for i := 0; i < DcCorNum; i++ {
		dcCor[i].Num = i
		rt1, rt2 := -1, -1 // 随机选择两个活动
		for {
			rt1 = util.RandomNumber(0, ProcessNum-1)*TaskNumPro + util.RandomNumber(0, TaskNumPro-1)
			rt2 = util.RandomNumber(0, ProcessNum-1)*TaskNumPro + util.RandomNumber(0, TaskNumPro-1)
			if rt1 == rt2 {
				break
			}
			s1 := rt1*SerNumPtask + util.RandomNumber(0, SerNumPtask-1) // 具有DC 关联的服务
			s2 := rt1*SerNumPtask + util.RandomNumber(0, SerNumPtask-1)
			dcCor[i].S1 = s1
			dcCor[i].S2 = s2
			dcCor[i].DcType = util.RandomNumber(0, 0)
			dcCor[i].Flag = true
		}
	}
	ActNum = ProcessNum * TaskNumPro //总活动数目=processNum*taskNumPro
	CpTask = 2

	for i := 0; i < 50; i++ {
		exeState.SerNum[i] = -1
	}

	runtimeState = true
	iniExePlan.GenBasicSolution(ProcessNum, TaskNumPro)
}

func serviceAttributeDefinition() {
	// 根据问题插入服务对应的活动编号，即b值，对多流程来说，编号为前i个流程*各流程活动数+第i个流程的第k个活动
	for p := 0; p < ProcessNum; p++ {
		for k := 0; k < TaskNumPro; k++ {
			for j := 0; j < SerNumPtask; j++ {
				servie[p*TaskNumPro*SerNumPtask+k*SerNumPtask+j].B = p*TaskNumPro + k
				servie[p*TaskNumPro*SerNumPtask+k*SerNumPtask+j].Seq = j
			}
		}
	}

	// 计算最大最小流程值
	calMQuality()

	cc := 0
	n := SerNumPtask * ProcessNum
	for i := 0; i < n; i++ {
		for {
			if cc >= ProcessNum*SerNumPtask {
				break
			}
			if servie[i].B == servie[n+cc].B {
				for k := 0; k < NrObj; k++ {
					if Obj[k].ObjType == 0 {
						if servie[n+cc].QosMM[k]-servie[n+cc].Qos[k] > 0 {
							servie[i].QosMM[k] = (servie[n+cc].QosMM[k] - servie[i].Qos[k]) / (servie[n+cc].QosMM[k] - servie[n+cc].Qos[k])
						} else {
							servie[i].QosMM[k] = 1
						}
					} else {
						if servie[n+cc].QosMM[k]-servie[n+cc].Qos[k] > 0 {
							servie[i].QosMM[k] = (servie[i].Qos[k] - servie[n+cc].QosMM[k]) / (servie[n+cc].QosMM[k] - servie[n+cc].Qos[k])
						} else {
							servie[i].QosMM[k] = 1
						}
					}
				}
				break
			}
			cc++
		}
	}

	for i := 0; i < ProcessNum*TaskNumPro; i++ {
		var ind []Service
		for j := 0; j < SerNumPtask; j++ {
			ind = append(ind, servie[i*SerNumPtask+j])
		}
		rankNum := Bm.PartitionIntoRanksService(ind)
		for j := 0; j < SerNumPtask; j++ {
			if rankNum == 1 {
				servie[i*SerNumPtask+j].F = 1.1
			} else {
				servie[i*SerNumPtask+j].F = 1.1 - (1.1-1)*(servie[i*SerNumPtask+j].F-1)/(float64(rankNum)-1)
			}
		}
	}
}

// 计算第i属性的最大、最小质量值
func calMQuality() {
	n := TaskNumPro * SerNumPtask * ProcessNum

	// 初始化n到n + processNum * func服务，用于存储各个活动的最小服务qos、最大服务qosMM
	for i := n; i < n+ProcessNum*TaskNumPro; i++ {
		servie[i].B = i - n
		servie[i].Qos = make([]float64, NrObj)
		servie[i].QosMM = make([]float64, NrObj)
		// 初始化为各活动的起始服务的QoS
		for k := 0; k < NrObj; k++ {
			servie[i].Qos[k] = servie[(i-n)*SerNumPtask].Qos[k]
			servie[i].QosMM[k] = servie[(i-n)*SerNumPtask].Qos[k]
		}
	}

	// 自候选的n个服务以后，存储了每个变迁中各属性的最好、最坏值
	cc := 0
	for i := 0; i < n; i++ {
		if servie[n+cc].B == servie[i].B {
			for k := 0; k < NrObj; k++ {
				if servie[i].Qos[k] < servie[n+cc].Qos[k] {
					servie[n+cc].Qos[k] = servie[i].Qos[k]
				}
				if servie[i].Qos[k] > servie[n+cc].QosMM[k] {
					servie[n+cc].QosMM[k] = servie[i].Qos[k]
				}
			}
		} else {
			cc++
		}
	}

	// 计算最好、最坏服务链组合
	for k := 0; k < NrObj; k++ {
		if Obj[k].AggreType_inPro == 0 {
			qualMinMax[k][0] = 0
			qualMinMax[k][1] = 0
		} else if Obj[k].AggreType_inPro == 1 {
			qualMinMax[k][0] = 1000000000
			qualMinMax[k][1] = 1000000000
		}
	}

	for p := 0; p < ProcessNum; p++ {
		// 计算子流程p的最大、最小目标解
		qsubPMin := make([]float64, NrObj) // 子流程p的最小目标解
		qsubPMax := make([]float64, NrObj) // 子流程p的最大目标解
		for k := 0; k < NrObj; k++ {
			if Obj[k].AggreType_inPro == 0 {
				qsubPMin[k] = 0
				qsubPMax[k] = 0
			} else if Obj[k].AggreType_inPro == 1 {
				qsubPMin[k] = 1
				qsubPMax[k] = 1
			}
		}

		for i := 0; i < TaskNumPro; i++ {
			v := n + p*TaskNumPro + i
			for k := 0; k < NrObj; k++ {
				qsubPMin[k] = AggQos(qsubPMin[k], Obj[k].AggreType_inPro, servie[v].Qos[k])
				qsubPMax[k] = AggQos(qsubPMax[k], Obj[k].AggreType_inPro, servie[v].QosMM[k])
			}
		}

		// 统计全局最大、最小流程解，并保存到qualMinMax
		for k := 0; k < NrObj; k++ {
			qualMinMax[k][0] = AggQos(qualMinMax[k][0], Obj[k].AggreType_betPro, qsubPMin[k])
			qualMinMax[k][1] = AggQos(qualMinMax[k][1], Obj[k].AggreType_betPro, qsubPMax[k])
		}
	}
}

// ===========Form4===========
var runtimeFlag bool
var runtimeState bool
var iniExePlan = BasicSolution{}

type EsStruct struct {
	ActNum []int
	SerNum []int
	RunNum int // 已执行的活动个数
}

var exeState = EsStruct{
	ActNum: []int{20},
	SerNum: make([]int, 50),
	RunNum: 1,
}
