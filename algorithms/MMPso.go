package algorithms

import (
	"Multi-objectiveOptimization/basic_class"
	"time"
)

type MMPso struct {
	PopSize            int // 种群规模
	ParallelFunc       int
	TotalFunc          int // 总活动数量
	CanCicyNum         int
	Generation         int // 遍历次数
	ConstraintsFitness basic_class.ConstraintsFitness
	BasicMooFunc       basic_class.BasicMooFunc
	MoeaqiBt           basic_class.MoeaqiBt
	StartTime          time.Time     // 开始时刻
	EndTime            time.Time     // 结束时刻
	TimeSpan           time.Duration // 时间段
	TimeConsume        float64
	ArrResult          []basic_class.BasicSolution
	CanSerAPro         [][]int
	NotSolNum          int       // 找不到的优解的个数
	IdealPoint         []float64 // new目标个数
	NarPoint           []float64
	Exa                []basic_class.BasicSolution // 存储外部最优种群，输出解集
	PBest              []basic_class.BasicSolution // 历史最优
	GBest              basic_class.BasicSolution   // 全局最优
	PSet               []basic_class.BasicSolution // 初始解集
	W                  float64
	C1                 float64
	C2                 float64
	C3                 float64
	CorFlag            bool
	PenFlag            bool
}

// 初始化
func (self *MMPso) Init(popSize int, totalFunc int, parallelFunc int, generation int, w float64, c1 float64, c2 float64, c3 float64) {
	self.MoeaqiBt = *new(basic_class.MoeaqiBt)                     // todo
	self.BasicMooFunc = *new(basic_class.BasicMooFunc)             //todo
	self.ConstraintsFitness = *new(basic_class.ConstraintsFitness) //todo
	self.PopSize = popSize
	self.TotalFunc = totalFunc
	self.ParallelFunc = parallelFunc
	self.Generation = generation
	self.W = w
	self.C1 = c1
	self.C2 = c2
	self.C3 = c3
	self.CorFlag = false
	self.PenFlag = false
}

func (self *MMPso) Run() {
	self.StartTime = time.Now()
	self.CanSerAPro = [][]int{} // todo
	self.PSet = self.randomPathinitial()
}

func (self *MMPso) randomPathinitial() []basic_class.BasicSolution {
	var pop []basic_class.BasicSolution

	// 产生初始解
	for i := 0; i < self.PopSize; i++ {
		// tempPath := new(basic_class.BasicSolution)
		// workNum := 0

	}
	return pop
}
