package basic_class

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
	ActNum: nil,
	SerNum: nil,
	RunNum: 0,
}
