package basic_class

import "math"

type BasicSolution struct {
	Solution      []int     // path1 存储所选服务
	Objective     []float64 // 存储目标值（适应度）
	Rank          int       // 该解的pareto排序号
	Sparsity      float64   // 该解的拥挤度
	TotalFit      float64   // SPEA2用
	Strength      float64   // SPEA2用 表示支配个数
	KthNNDistance float64   // SPEA2用
	A             float64
	ASol          int
	ASF           float64
	ASFRank       int
	CtnValue      []float64
	Vel           []float64
	VelType       int
	Pos           []float64
	Order         int       // 记录解编号
	Num           int       // 解编号
	V             []float64 // 速度 PSO方法用
	X             []float64 // 位置 PSO方法用
	SubProbNo     int       // 存储解被分配到的权重向量
	TchVal        float64   // 存储PBI值
	FitnessValue  float64
	Trail         int     // MS-DABC used 表示蜜蜂被跟随的概率
	Probability   float64 // MS-DABC used 表示蜜蜂被跟随的概率
	Angle         float64
	RealGenes     []float64 // decision variable
	STime         []float64 // 各活动的开始时间
	Selected      bool
}

func (self *BasicSolution) GenBasicSolution(processNum int, taskNumPro int) {
	for i := 0; i < processNum*taskNumPro; i++ {
		self.Solution = append(self.Solution, 0)
		self.CtnValue = append(self.CtnValue, 0.0)
		self.Vel = append(self.Vel, 0.0)
		self.Pos = append(self.CtnValue, 0.0)
		self.V = append(self.CtnValue, 0.0)
		self.X = append(self.CtnValue, 0.0)
		self.STime = append(self.STime, 0.0)
	}
	self.Objective = []float64{}
}

func Copy(fromSol BasicSolution, toSol *BasicSolution) {
	toSol.Solution = make([]int, len(fromSol.Solution))
	toSol.X = make([]float64, len(fromSol.X))
	toSol.V = make([]float64, len(fromSol.V))
	for i := 0; i < len(fromSol.Solution); i++ {
		toSol.Solution[i] = fromSol.Solution[i]
		toSol.X[i] = fromSol.X[i]
		toSol.V[i] = fromSol.V[i]
	}
	toSol.Objective = []float64{}
	for i := 0; i < len(fromSol.Objective); i++ {
		toSol.Objective = append(toSol.Objective, fromSol.Objective[i])
	}
	toSol.Rank = fromSol.Rank
	toSol.Sparsity = fromSol.Sparsity
	toSol.TotalFit = fromSol.TotalFit
	toSol.Strength = fromSol.Strength
	toSol.KthNNDistance = fromSol.KthNNDistance
}

func (self *BasicSolution) GetRank() int {
	return self.Rank
}

func (self *BasicSolution) SetRank(rank int) {
	self.Rank = rank
}

func (self *BasicSolution) CopyTo(copyto *BasicSolution) {
	copyto.FitnessValue = self.FitnessValue
	copyto.TchVal = self.TchVal
	copyto.SubProbNo = self.SubProbNo
	copyto.Trail = self.Trail
	copyto.Probability = self.Probability
	copyto.Angle = self.Angle
	copyto.Rank = self.Rank

	for i := 0; i < len(copyto.Solution); i++ {
		copyto.Solution[i] = self.Solution[i]
	}

	for i := 0; i < len(self.Objective); i++ {
		copyto.Objective = append(copyto.Objective, self.Objective[i])
	}
}

func (self *BasicSolution) Dominates(another BasicSolution) bool {
	flag := true
	sNum := 0
	// 若各目标值相同，则不能算支配
	for i := 0; i < NrObj; i++ {
		if math.Abs(self.Objective[i]-another.Objective[i]) < 0.0000001 {
			sNum++
		}
	}
	if sNum == NrObj {
		return false
	}

	for i := 0; i < NrObj; i++ {
		if self.Objective[i] > another.Objective[i] {
			flag = false
			return flag
		}
	}
	return flag
}
