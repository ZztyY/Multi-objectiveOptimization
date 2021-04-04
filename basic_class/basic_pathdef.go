package basic_class

type BasicPathDef struct {
	Num   int   // 解编号
	Path1 []int // path1存储所选服务
	// Objective []float64 Objective // 多个目标值
	Objective      []float64
	ObjectiveStd   []float64 // MFF用
	ObjectiveStand []float64 // E3SLA_X3yong
	// Vel []float64
	Rank     int       // 该解的帕累托排序号
	Sparsity float64   // 该解的拥挤度
	TotalFit float64   // SPEA2 用
	Order    int       // 记录解编号
	STime    []float64 // 各活动开始时间
	CtnValue []float64 // path1对应的连续优化问题值 // 存储连续值，连续值取整后对应path值
	Vel      []float64
	Pos      []float64
	Q        []float64
	V        []float64 // 速度
	Sol      []float64 // 位置
}

func (self *BasicPathDef) BasicPathDef() {
	self.Path1 = make([]int, ActNum)
	self.CtnValue = make([]float64, ActNum)
	self.Vel = make([]float64, ActNum)
	self.Pos = make([]float64, ActNum)
	self.Q = make([]float64, ActNum)
	self.V = make([]float64, ActNum)
	self.Sol = make([]float64, ActNum)
}
