package MOEAs

import (
	"Multi-objectiveOptimization/basic_class"
	"Multi-objectiveOptimization/util"
	"math"
	"math/rand"
	"time"
)

type NSGA_3 struct {
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
	TimeConsume        int64
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
	NumRanks           int // 非支配层的数量
	Weights            [][]float64
	Div                int
	NeighbourSize      int
	NeighbourTable     [][]int
	MainPop            []basic_class.BasicSolution
	TMinusOnePop       []basic_class.BasicSolution
	TMinusTwoPop       []basic_class.BasicSolution
	SubregionIdx_      [][]int // index matrix for subregion record，subregionIdx_[i,j]=1表示EXA中的第j个解属于权重向量i
	RankIdx_           [][]int // index matrix for the non-domination levels，rankIdx_[i,j]=1表示EXA中的第j个解属于第i个perato层
	Nr                 int
	ItrCounter         int
	TotalItrNum        int
	IsNormalization    bool
}

// 初始化
func (self *NSGA_3) Init(popSize int, totalFunc int, parallelFunc int, generation int, w float64, c1 float64, c2 float64, c3 float64) {
	self.MoeaqiBt = *new(basic_class.MoeaqiBt)
	self.BasicMooFunc = *new(basic_class.BasicMooFunc)
	self.ConstraintsFitness = *new(basic_class.ConstraintsFitness)
	self.PopSize = popSize
	self.TotalFunc = totalFunc
	self.ParallelFunc = parallelFunc
	self.Generation = generation
	self.IdealPoint = make([]float64, basic_class.NrObj)
	self.NarPoint = make([]float64, basic_class.NrObj)
	self.GBest = *new(basic_class.BasicSolution)
	self.GBest.GenBasicSolution(basic_class.ProcessNum, basic_class.TaskNumPro)
	self.W = w
	self.C1 = c1
	self.C2 = c2
	self.C3 = c3
	self.CorFlag = true
	self.PenFlag = true
	self.Nr = 2
	self.ItrCounter = 1
	self.TotalItrNum = 50
	self.IsNormalization = false
	if basic_class.NrObj <= 2 {
		self.Div = 69
	} else {
		self.Div = 5
	}
	self.NeighbourSize = 20
	for i := 0; i < basic_class.NrObj; i++ {
		self.IdealPoint[i] = 1000000.0
		self.NarPoint[i] = 0.0
	}
	self.CanSerAPro = self.MoeaqiBt.ServicePreProcess() // 服务预处理
	self.InitWeight(self.Div)                           // 计算初始权重向量，并存储到weights
	self.NeighbourSize = 5
	self.InitNeighbour() // 计算初始权重向量的邻居，并存储到neighbourTable

	self.InitialPopulation() // 生成初始解集合mainpop，其mainpop[i]存储了第i个权重向量对应的解

	// 计算每个解的所属区域
	self.SubregionIdx_ = make([][]int, self.PopSize+1) // 用于存储每个权重向量所包含的解：subregionIdx_[i][j]=1表示EXA中的第j个解属于权重向量i
	for k, _ := range self.SubregionIdx_ {
		self.SubregionIdx_[k] = make([]int, self.PopSize+1)
	}
	for k := 0; k < self.PopSize; k++ {
		self.SetLocation(&self.MainPop[k])                                                            // 存储到offSpring.subProbNo，offSpring.angle
		self.MainPop[k].TchVal = self.PbiScalarObj(self.MainPop[k].SubProbNo, self.MainPop[k], false) // 计算PBI值
		self.SubregionIdx_[self.MainPop[k].SubProbNo][k] = 1                                          // 存储EXA集合中每个权重向量包含了哪些解
	}

	// 确定解所在的非支配层
	self.RankIdx_ = make([][]int, self.PopSize+1) // 用于存储各解属于哪个层：rankIdx_[i][j]=1表示EXA中的第j个解属于第i个perato层
	for k, _ := range self.RankIdx_ {
		self.RankIdx_[k] = make([]int, self.PopSize+1)
	}
	self.FastNonDominatedSort(self.MainPop) // 根据非支配排序法确定各个解所在的层数,可获得List<List<Basic_Pathdef>> frontSet集合及更新EXA[i].rank的值
	var curRank int
	for i := 0; i < self.PopSize; i++ {
		curRank = self.MainPop[i].GetRank()
		self.RankIdx_[curRank][i] = 1 // 存储EXA集合中排到第i个pareto层包含了哪些解
	}

	// 复制MainPop到EXA
	for i := 0; i < len(self.MainPop); i++ {
		var basicPathdef basic_class.BasicSolution
		self.MainPop[i].CopyTo(&basicPathdef)
		self.Exa = append(self.Exa, basicPathdef)
	}
}

func (self *NSGA_3) InitWeight(m int) {
	var uniPointGenerator basic_class.UniPointsGenerator
	if basic_class.NrObj < 5 {
		self.Weights = uniPointGenerator.GetMUniDistributedPoint(basic_class.NrObj, m)
	}
	self.PopSize = len(self.Weights)
}

// 初始化邻居
func (self *NSGA_3) InitNeighbour() {
	self.NeighbourTable = make([][]int, self.PopSize) // 用一个种群规模大小的列表来存储邻居

	distanceMatrix := make([][]float64, self.PopSize) // 用矩阵来记录权重向量i与j之间的距离
	for k, _ := range distanceMatrix {
		distanceMatrix[k] = make([]float64, self.PopSize)
	}
	for i := 0; i < self.PopSize; i++ {
		distanceMatrix[i][i] = 0
		for j := i + 1; j < self.PopSize; j++ {
			distanceMatrix[i][j] = Distance(self.Weights[i], self.Weights[j]) // 计算欧氏距离
			distanceMatrix[j][i] = distanceMatrix[i][j]
		}
	}

	for i := 0; i < self.PopSize; i++ {
		val := make([]float64, self.PopSize) // 一个种群规模大小的数组
		for j := 0; j < self.PopSize; j++ {
			val[j] = distanceMatrix[i][j] // 存储i与j距离的值
		}

		index := basic_class.Sort(val)
		array := make([]int, self.NeighbourSize)
		for k := 0; k < self.NeighbourSize; k++ {
			array[k] = index[k]
		} // todo
		self.NeighbourTable[i] = array
	}
}

func (self *NSGA_3) InitialPopulation() {
	for {
		if len(self.MainPop) >= self.PopSize {
			break
		}
		basicPathdef := self.CreateBasicSolution()
		// 计算适应度
		{
			// 保存到解集
			self.UpdateReference(basicPathdef) // 更新最好解集
			self.MainPop = append(self.MainPop, basicPathdef)
		}
	}
}

func (self *NSGA_3) CreateBasicSolution() basic_class.BasicSolution {
	var basicPathdef basic_class.BasicSolution
	basicPathdef.Solution = make([]int, basic_class.ProcessNum*basic_class.TaskNumPro)
	basicPathdef.Objective = *new([]float64)

	workNum := 0
	for p := 0; p < basic_class.ProcessNum; p++ { // 转换成适应度，计算适应度总和
		for j := 0; j < basic_class.TaskNumPro; j++ {
			nextCust := util.RandomNumber(0, len(self.CanSerAPro[workNum])-1)
			basicPathdef.Solution[workNum] = self.CanSerAPro[workNum][nextCust]
			workNum++
		}
	}

	// 计算适应度
	basicPathdef.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(basicPathdef.Solution, self.CorFlag, self.PenFlag) // 计算适应度值
	return basicPathdef
}

func Distance(weight1 []float64, weight2 []float64) float64 {
	sum := 0.0
	for i := 0; i < len(weight1); i++ {
		sum += math.Pow((weight1[i] - weight2[i]), 2)
	}
	return math.Sqrt(sum)
}

func (self *NSGA_3) SetLocation(offspring *basic_class.BasicSolution) {
	theta := 100000000.0
	idx := 0
	var ta float64
	for i := 0; i < len(self.Weights); i++ {
		ta = self.GetAngle(i, *offspring, false)
		if ta < theta {
			theta = ta
			idx = i
		}
	}
	offspring.SubProbNo = idx
	offspring.Angle = theta
}

func (self *NSGA_3) GetAngle(idx int, v basic_class.BasicSolution, flag bool) float64 {
	namda := self.Weights[idx]
	mul := 0.0
	a := 0.0
	b := 0.0
	for i := 0; i < basic_class.NrObj; i++ {
		if flag {
			mul += ((v.Objective[i] - self.IdealPoint[i] + 1e-5) / (self.NarPoint[i] - self.IdealPoint[i] + 1e-5)) * namda[i] // 归一
			b += math.Pow((v.Objective[i]-self.IdealPoint[i]+1e-5)/(self.NarPoint[i]-self.IdealPoint[i]+1e-5), 2)             // 所有归一值的总和
		} else {
			mul += namda[i] * (v.Objective[i] - self.IdealPoint[i])
			b += math.Pow(v.Objective[i]-self.IdealPoint[i]+1e-5, 2) // 归一化后的直边,解与idealpoint的距离
		}
		a += math.Pow(namda[i], 2) // 斜边，参考点与原点的距离
	}
	return math.Acos(mul / (math.Sqrt(a * b)))
}

func (self *NSGA_3) PbiScalarObj(idx int, v basic_class.BasicSolution, flag bool) float64 {
	// 计算d1
	namda := self.Weights[idx]
	lenv := 0.0
	mul := 0.0
	for i := 0; i < basic_class.NrObj; i++ {
		if flag {
			mul += ((v.Objective[i] - self.IdealPoint[i]) / (self.NarPoint[i] - self.IdealPoint[i] + 1e-5)) * namda[i]
		} else {
			mul += (v.Objective[i] - self.IdealPoint[i]) * namda[i]
		}
		lenv += math.Pow(namda[i], 2)
	}
	d1 := mul / math.Sqrt(lenv) // 当目标值离理想点越远，d1越大

	d2 := 0.0
	for i := 0; i < basic_class.NrObj; i++ {
		if flag {
			d2 += math.Pow(((v.Objective[i]-self.IdealPoint[i])/(self.NarPoint[i]-self.IdealPoint[i]+1e-5))-d1*namda[i]/math.Sqrt(lenv), 2)
		} else {
			d2 += math.Pow(v.Objective[i]-self.IdealPoint[i]-d1*namda[i]/math.Sqrt(lenv), 2)
		}
	}
	return d1 + 5*math.Sqrt(d2)
}

func (self *NSGA_3) FastNonDominatedSort(individuals []basic_class.BasicSolution) [][]basic_class.BasicSolution {
	var dominationFronts [][]int
	individual2DominatedIndividuals := map[int][]int{}
	individual2NumberOfDominatingIndividuals := map[int]int{}

	s := 0
	t := 0
	for _, individualP := range individuals {
		individual2DominatedIndividuals[s] = []int{}
		individual2NumberOfDominatingIndividuals[s] = 0

		t = 0
		for _, individualQ := range individuals {
			if individualP.Dominates(individualQ) {
				individual2DominatedIndividuals[s] = append(individual2DominatedIndividuals[s], t)
			} else {
				if individualQ.Dominates(individualP) {
					individual2NumberOfDominatingIndividuals[s] = individual2NumberOfDominatingIndividuals[s] + 1
				}
			}
			t++
		}

		if individual2NumberOfDominatingIndividuals[s] == 0 {
			// p belongs to the first front
			individualP.SetRank(0)
			if len(dominationFronts) == 0 {
				var firstDominationFront []int
				firstDominationFront = append(firstDominationFront, s)
				dominationFronts = append(dominationFronts, firstDominationFront)
			} else {
				dominationFronts[0] = append(dominationFronts[0], s)
			}
		}
		s++
	}

	i := 1
	for {
		var nextDominationFront []int
		for _, individualP := range dominationFronts[i-1] {
			for _, individualQ := range individual2DominatedIndividuals[individualP] {
				individual2NumberOfDominatingIndividuals[individualQ] = individual2NumberOfDominatingIndividuals[individualQ] - 1
				if individual2NumberOfDominatingIndividuals[individualQ] == 0 {
					individuals[individualQ].SetRank(i)
					nextDominationFront = append(nextDominationFront, individualQ)
				}
			}
		}
		i++
		if len(nextDominationFront) != 0 {
			dominationFronts = append(dominationFronts, nextDominationFront)
		} else {
			break
		}
	}

	var frontSet [][]basic_class.BasicSolution
	for x := 0; x < len(dominationFronts); x++ {
		var it []basic_class.BasicSolution
		for v := 0; v < len(dominationFronts[x]); v++ {
			individuals[dominationFronts[x][v]].Rank = x
			it = append(it, individuals[dominationFronts[x][v]])
		}
		frontSet = append(frontSet, it)
	}
	return frontSet
}

func (self *NSGA_3) Terminated() bool {
	return self.ItrCounter > self.TotalItrNum
}

func (self *NSGA_3) Run() {
	self.StartTime = time.Now()
	self.TMinusOnePop = self.MainPop
	self.TMinusTwoPop = self.MainPop
	for {
		if self.Terminated() {
			break
		}
		var offsPop []basic_class.BasicSolution

		for i := 0; i < self.PopSize; i++ {
			offspring := self.GenNewPop_1(i)
			offsPop = append(offsPop, offspring)
			self.UpdateReference(offspring)
		}

		// 六种信息反馈模型
		//offsPop = self.MF1(offsPop)
		//offsPop = self.MR1(offsPop)
		//offsPop = self.MF2(offsPop)
		//offsPop = self.MR2(offsPop)
		//offsPop = self.MF3(offsPop)
		//offsPop = self.MR3(offsPop)

		var Pop []basic_class.BasicSolution
		for k, _ := range self.MainPop {
			Pop = append(Pop, self.MainPop[k])
		}
		for k, _ := range offsPop {
			Pop = append(Pop, offsPop[k])
		}

		self.EnviromentSelection(Pop)

		self.ItrCounter++
	}

	//=============MS-DABC============
	//for q := 0; q < self.Generation; q++ {
	//	for i := 0; i < self.PopSize; i++ {
	//		//Step1：从权重向量i的邻域中选出父代解，使用GA生成子解，并计算适应度
	//		offspring := self.GenNewPop_1(i)
	//
	//		//Step2：更新ideal point、narpoint
	//		self.UpdateReference(offspring)
	//
	//		//Step3：计算offspring属于哪个权重向量，并存储其PBI值
	//		self.SetLocation(offspring)                                                 // 存入offspring.subProbNo，offspring.angle
	//		offspring.TchVal = self.PbiScalarObj(offspring.SubProbNo, offspring, false) // 计算PBI值
	//
	//		//Step4:根据所属的权重向量，更新该权重向量对应的解及邻域解
	//		self.UpdateNeighbours_1(offspring) // 更新邻域解，即更新MainPop集合
	//
	//		//Step5：更新EXA集合
	//		self.UpdateArchive(offspring)
	//	}
	//}

	self.EndTime = time.Now()
	// self.TimeSpan = self.EndTime - self.StartTime todo
	self.TimeConsume = self.TimeSpan.Milliseconds()

	// 将归一化的目标值重新计算为QoS值
	//for i := 0; i < len(self.Exa); i++ {
	//	self.Exa[i].Objective = self.ConstraintsFitness.CalFitnessMoo(self.Exa[i].Solution, true, true)
	//	conCount := self.ConstraintsFitness.CalTotalConstraint(self.Exa[i].Solution, true, true, true)
	//	self.Exa[i].Objective = append(self.Exa[i].Objective, conCount)
	//}
	//
	//self.ArrResult = self.Exa
}

// 选择解
func (self *NSGA_3) SelectClose(canSerAPro []int, sol int) int {
	res := sol
	min := 10000
	for _, v := range canSerAPro {
		if v == sol {
			res = sol
			return res
		}
		if int(math.Abs(float64(v-sol))) < min {
			min = int(math.Abs(float64(v - sol)))
			res = v
		}
	}
	return res
}

// 信息反馈M-F1,k=i
func (self *NSGA_3) MF1(u []basic_class.BasicSolution) []basic_class.BasicSolution {
	self.CalPopFit(u)
	self.CalPopFit(self.MainPop)
	var res []basic_class.BasicSolution
	for k, _ := range u {
		var temp basic_class.BasicSolution
		a1 := self.MainPop[k].TotalFit / (u[k].TotalFit + self.MainPop[k].TotalFit)
		a2 := u[k].TotalFit / (u[k].TotalFit + self.MainPop[k].TotalFit)
		for j, _ := range u[k].Solution {
			temp.Solution = append(temp.Solution, self.SelectClose(self.CanSerAPro[j], int(a1*float64(u[k].Solution[j])+a2*float64(self.MainPop[k].Solution[j]))))
		}
		temp.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(temp.Solution, self.CorFlag, self.PenFlag)
		res = append(res, temp)
	}
	return res
}

// 信息反馈M-R1
func (self *NSGA_3) MR1(u []basic_class.BasicSolution) []basic_class.BasicSolution {
	self.CalPopFit(u)
	self.CalPopFit(self.MainPop)
	var res []basic_class.BasicSolution
	for k, _ := range u {
		var temp basic_class.BasicSolution
		m := util.RandomNumber(0, len(self.MainPop)-1) // 此处m为公式中的k,随机整数生成区别于M-F1
		a1 := self.MainPop[m].TotalFit / (u[k].TotalFit + self.MainPop[m].TotalFit)
		a2 := u[k].TotalFit / (u[k].TotalFit + self.MainPop[m].TotalFit)
		for j, _ := range u[k].Solution {
			temp.Solution = append(temp.Solution, self.SelectClose(self.CanSerAPro[j], int(a1*float64(u[k].Solution[j])+a2*float64(self.MainPop[m].Solution[j]))))
		}
		temp.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(temp.Solution, self.CorFlag, self.PenFlag)
		res = append(res, temp)
	}
	return res
}

// 信息反馈M-F2
func (self *NSGA_3) MF2(u []basic_class.BasicSolution) []basic_class.BasicSolution {
	self.CalPopFit(u)
	self.CalPopFit(self.MainPop)
	var res []basic_class.BasicSolution
	for k, _ := range u {
		var temp basic_class.BasicSolution
		a1 := (self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit) / (2 * (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit))
		a2 := (u[k].TotalFit + self.TMinusOnePop[k].TotalFit) / (2 * (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit))
		a3 := (u[k].TotalFit + self.MainPop[k].TotalFit) / (2 * (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit))
		for j, _ := range u[k].Solution {
			temp.Solution = append(temp.Solution, self.SelectClose(self.CanSerAPro[j], int(a1*float64(u[k].Solution[j])+a2*float64(self.MainPop[k].Solution[j])+a3*float64(self.TMinusOnePop[k].Solution[j]))))
		}
		temp.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(temp.Solution, self.CorFlag, self.PenFlag)
		res = append(res, temp)
	}
	self.TMinusOnePop = self.MainPop
	return res
}

func (self *NSGA_3) MR2(u []basic_class.BasicSolution) []basic_class.BasicSolution {
	self.CalPopFit(u)
	self.CalPopFit(self.MainPop)
	var res []basic_class.BasicSolution
	for k, _ := range u {
		var temp basic_class.BasicSolution
		m := util.RandomNumber(0, len(self.MainPop)-1) // 此处m为公式中的k,随机整数生成区别于M-F2
		a1 := (self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit) / (2 * (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit))
		a2 := (u[k].TotalFit + self.TMinusOnePop[m].TotalFit) / (2 * (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit))
		a3 := (u[k].TotalFit + self.MainPop[m].TotalFit) / (2 * (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit))
		for j, _ := range u[k].Solution {
			temp.Solution = append(temp.Solution, self.SelectClose(self.CanSerAPro[j], int(a1*float64(u[k].Solution[j])+a2*float64(self.MainPop[m].Solution[j])+a3*float64(self.TMinusOnePop[m].Solution[j]))))
		}
		temp.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(temp.Solution, self.CorFlag, self.PenFlag)
		res = append(res, temp)
	}
	self.TMinusOnePop = self.MainPop
	return res
}

func (self *NSGA_3) MF3(u []basic_class.BasicSolution) []basic_class.BasicSolution {
	self.CalPopFit(u)
	self.CalPopFit(self.MainPop)
	var res []basic_class.BasicSolution
	for k, _ := range u {
		var temp basic_class.BasicSolution
		a1 := (self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit + self.TMinusTwoPop[k].TotalFit) / (3 * (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit + self.TMinusTwoPop[k].TotalFit))
		a2 := (u[k].TotalFit + self.TMinusOnePop[k].TotalFit + self.TMinusTwoPop[k].TotalFit) / (3 * (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit + self.TMinusTwoPop[k].TotalFit))
		a3 := (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusTwoPop[k].TotalFit) / (3 * (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit + self.TMinusTwoPop[k].TotalFit))
		a4 := (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit) / (3 * (u[k].TotalFit + self.MainPop[k].TotalFit + self.TMinusOnePop[k].TotalFit + self.TMinusTwoPop[k].TotalFit))
		for j, _ := range u[k].Solution {
			temp.Solution = append(temp.Solution, self.SelectClose(self.CanSerAPro[j], int(a1*float64(u[k].Solution[j])+a2*float64(self.MainPop[k].Solution[j])+a3*float64(self.TMinusOnePop[k].Solution[j])+a4*float64(self.TMinusTwoPop[k].Solution[j]))))
		}
		temp.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(temp.Solution, self.CorFlag, self.PenFlag)
		res = append(res, temp)
	}
	self.TMinusTwoPop = self.TMinusOnePop
	self.TMinusOnePop = self.MainPop
	return res
}

func (self *NSGA_3) MR3(u []basic_class.BasicSolution) []basic_class.BasicSolution {
	self.CalPopFit(u)
	self.CalPopFit(self.MainPop)
	var res []basic_class.BasicSolution
	for k, _ := range u {
		var temp basic_class.BasicSolution
		m := util.RandomNumber(0, len(self.MainPop)-1) // 此处m为公式中的k,随机整数生成区别于M-F2
		a1 := (self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit + self.TMinusTwoPop[m].TotalFit) / (3 * (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit + self.TMinusTwoPop[m].TotalFit))
		a2 := (u[k].TotalFit + self.TMinusOnePop[m].TotalFit + self.TMinusTwoPop[m].TotalFit) / (3 * (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit + self.TMinusTwoPop[m].TotalFit))
		a3 := (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusTwoPop[m].TotalFit) / (3 * (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit + self.TMinusTwoPop[m].TotalFit))
		a4 := (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit) / (3 * (u[k].TotalFit + self.MainPop[m].TotalFit + self.TMinusOnePop[m].TotalFit + self.TMinusTwoPop[m].TotalFit))
		for j, _ := range u[k].Solution {
			temp.Solution = append(temp.Solution, self.SelectClose(self.CanSerAPro[j], int(a1*float64(u[k].Solution[j])+a2*float64(self.MainPop[m].Solution[j])+a3*float64(self.TMinusOnePop[m].Solution[j])+a4*float64(self.TMinusTwoPop[m].Solution[j]))))
		}
		temp.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(temp.Solution, self.CorFlag, self.PenFlag)
		res = append(res, temp)
	}
	self.TMinusTwoPop = self.TMinusOnePop
	self.TMinusOnePop = self.MainPop
	return res
}

func (self *NSGA_3) EnviromentSelection(pop []basic_class.BasicSolution) {
	var result []basic_class.BasicSolution
	dominatedSet0 := self.FastNonDominatedSort(pop)
	if self.IsNormalization { //
		self.UpdateNadirPoint(dominatedSet0[0])
	}

	cnt := 0
	for {
		if len(result)+len(dominatedSet0[cnt]) > self.PopSize {
			break
		}
		for r := 0; r < len(dominatedSet0[cnt]); r++ {
			dominatedSet0[cnt][r].Selected = true
		}

		for k, _ := range dominatedSet0[cnt] {
			result = append(result, dominatedSet0[cnt][k])
		}
		cnt++
	}
	if len(result) == self.PopSize {
		self.MainPop = *new([]basic_class.BasicSolution)
		for k, _ := range result {
			self.MainPop = append(self.MainPop, result[k])
		}
		return
	}

	count := make([]int, self.PopSize)
	for i := 0; i < self.PopSize; i++ {
		count[i] = 0
	}

	for i := 0; i < len(result); i++ {
		dist := 100000000.0
		var dt float64
		pos := -1
		for j := 0; j < len(self.Weights); j++ {
			dt = self.GetAngle(j, result[i], self.IsNormalization)
			if dt < dist {
				dist = dt
				pos = j
			}
		}
		count[pos]++
	}

	var associatedSolution [][]basic_class.BasicSolution
	for i := 0; i < len(self.Weights); i++ {
		var temp []basic_class.BasicSolution
		associatedSolution = append(associatedSolution, temp)
	}

	for i := 0; i < len(dominatedSet0[cnt]); i++ {
		dist := 100000000.0
		var dt float64
		pos := -1
		for j := 0; j < len(self.Weights); j++ {
			dt = self.GetAngle(j, dominatedSet0[cnt][i], self.IsNormalization)
			if dt < dist {
				dist = dt
				pos = j
			}
		}
		dominatedSet0[cnt][i].TchVal = self.PbiScalarObj(pos, dominatedSet0[cnt][i], self.IsNormalization)
		dominatedSet0[cnt][i].SubProbNo = pos
		associatedSolution[pos] = append(associatedSolution[pos], dominatedSet0[cnt][i])
	}

	for i := 0; i < len(self.Weights); i++ {
		for j := 0; j < len(associatedSolution[i]); j++ {
			for k := j + 1; k < len(associatedSolution[i]); k++ {
				if associatedSolution[i][j].TchVal < associatedSolution[i][k].TchVal {
					temp := associatedSolution[i][j]
					associatedSolution[i][j] = associatedSolution[i][k]
					associatedSolution[i][k] = temp
				}
			}
		}
	}

	var itr []int
	for {
		if len(result) >= self.PopSize {
			break
		}
		itr = *new([]int)
		min := self.PopSize
		for i := 0; i < self.PopSize; i++ {
			if min > count[i] && len(associatedSolution[i]) > 0 {
				itr = *new([]int)
				min = count[i]
				itr = append(itr, i)
			} else if min == count[i] && len(associatedSolution[i]) > 0 {
				itr = append(itr, i)
			}
		}

		pos := util.RandomNumber(0, len(itr)-1)
		count[itr[pos]]++
		result = append(result, associatedSolution[itr[pos]][0])
		associatedSolution[itr[pos]] = associatedSolution[itr[pos]][1:]
	}
	self.MainPop = *new([]basic_class.BasicSolution)
	for k, _ := range result {
		self.MainPop = append(self.MainPop, result[k])
	}
	return
}

func (self *NSGA_3) UpdateNadirPoint(list []basic_class.BasicSolution) {
	var li []int
	for i := 0; i < basic_class.NrObj; i++ {
		min := 100000000.0
		pos := -1

		w := make([]float64, basic_class.NrObj)
		for j := 0; j < basic_class.NrObj; j++ {
			if i == j {
				w[j] = 1
			} else {
				w[j] = 1e-6
			}
		}

		for j := 0; j < len(list); j++ {
			max := 0.0
			for r := 0; r < basic_class.NrObj; r++ {
				tp := (list[j].Objective[r] - self.IdealPoint[r]) / w[r]
				if max < tp {
					max = tp
				}
			}
			if min > max {
				min = max
				pos = j
			}
		}
		li = append(li, pos)
	}

	arr := make([][]float64, basic_class.NrObj)
	for k, _ := range arr {
		arr[k] = make([]float64, basic_class.NrObj)
	}
	for i := 0; i < basic_class.NrObj; i++ {
		for j := 0; j < basic_class.NrObj; j++ {
			arr[i][j] = list[li[i]].Objective[j] - self.IdealPoint[j]
		}
	}
	iMatrix := basic_class.IMatrix(arr, basic_class.NrObj)
	u := make([]float64, basic_class.NrObj)
	for i := 0; i < basic_class.NrObj; i++ {
		u[i] = 1
	}
	result := basic_class.MatrixMultiple(iMatrix, u)

	if result == nil || !IsSatisfy(result) {
		for i := 0; i < basic_class.NrObj; i++ {
			self.NarPoint[i] = 0.0
		}
		for i := 0; i < basic_class.NrObj; i++ {
			for j := 0; j < len(list); j++ {
				if list[j].Objective[i] > self.NarPoint[i] {
					self.NarPoint[i] = list[j].Objective[i]
				}
			}
		}
	} else {
		for i := 0; i < basic_class.NrObj; i++ {
			self.NarPoint[i] = 1.0/result[i] + self.IdealPoint[i]
		}
	}
}

func IsSatisfy(arr []float64) bool {
	flag := true
	for _, e := range arr {
		if e <= 1e-5 || math.NaN() == e {
			return false
		}
	}
	return flag
}

func (self *NSGA_3) randomPathinitial() []basic_class.BasicSolution {
	var pop []basic_class.BasicSolution

	// 产生初始解
	for i := 0; i < self.PopSize; i++ {
		// 在起始变迁中随机选择第一个客户
		tempPath := new(basic_class.BasicSolution)
		tempPath.GenBasicSolution(basic_class.ProcessNum, basic_class.TaskNumPro)
		workNum := 0
		for p := 0; p < basic_class.ProcessNum; p++ {
			for j := 0; j < basic_class.TaskNumPro; j++ {
				nextCust := util.RandomNumber(0, len(self.CanSerAPro[workNum])-1)
				tempPath.Solution[workNum] = self.CanSerAPro[workNum][nextCust]
				tempPath.X[workNum] = float64(tempPath.Solution[workNum])
				workNum++
			}
		}
		tempPath.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(tempPath.Solution, self.CorFlag, self.PenFlag)
		self.UpdateReference(*tempPath)

		pop = append(pop, *tempPath)
	}
	return pop
}

//在初始化阶段，更新EXA为初始解集中的最优解
func (self *NSGA_3) UpdateEXAIni(particleSet []basic_class.BasicSolution) {
	// 更新pbest
	for i := 0; i < self.PopSize; i++ {
		temp := new(basic_class.BasicSolution)
		temp.GenBasicSolution(basic_class.ProcessNum, basic_class.TaskNumPro)
		basic_class.Copy(particleSet[i], temp)
		self.PBest = append(self.PBest, *temp)
	}

	// 更新EXA
	var front []basic_class.BasicSolution
	front = self.FindFrontMinNoCon(particleSet)
	for i := 0; i < len(front); i++ {
		temp := new(basic_class.BasicSolution)
		temp.GenBasicSolution(basic_class.ProcessNum, basic_class.TaskNumPro)
		basic_class.Copy(front[i], temp)
		self.Exa = append(self.Exa, *temp)
	}

	// 更新gbest
	var selIndex int
	selIndex = util.RandomNumber(0, len(self.Exa)-1)
	basic_class.Copy(self.Exa[selIndex], &self.GBest)
}

func Randomfloat64() float64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Float64()
}

// PSO算法
func (self *NSGA_3) GenNewPopPSO(particleSet []basic_class.BasicSolution) []basic_class.BasicSolution {
	var offspringSet []basic_class.BasicSolution
	for i := 0; i < self.PopSize; i++ {
		var offspring basic_class.BasicSolution
		offspring.GenBasicSolution(basic_class.ProcessNum, basic_class.TaskNumPro)
		parent := particleSet[i]
		index := 0
		for p := 0; p < basic_class.ProcessNum; p++ {
			for j := 0; j < basic_class.TaskNumPro; j++ {
				offspring.V[index] = self.W*parent.V[index] + self.C1*Randomfloat64()*(self.PBest[i].X[index]-parent.X[index]) + self.C2*Randomfloat64()*(self.GBest.X[index]-parent.X[index]) + self.C3*Randomfloat64()*(self.GBest.X[index]-self.PBest[i].X[index])
				offspring.X[index] = parent.X[index] + offspring.V[index]
				// 转化成整数
				offspring.Solution[index] = self.DouToInt1(offspring.X[index], index) // 将连续值转化成有效的服务编号
				index++
			}
		}

		// 计算适应度
		offspring.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(offspring.Solution, self.CorFlag, self.PenFlag)
		offspringSet = append(offspringSet, offspring)

		// 更新最近点和最远点
		self.UpdateReference(offspring)

		// 更新pBest
		if self.BasicMooFunc.ParetoDominatesMin(offspring.Objective, self.PBest[i].Objective) {
			basic_class.Copy(offspring, &self.PBest[i])
		}
	}

	// 更新PSet，为下一次迭代用
	self.PSet = []basic_class.BasicSolution{}
	for i := 0; i < self.PopSize; i++ {
		temp := new(basic_class.BasicSolution)
		basic_class.Copy(offspringSet[i], temp)
		self.PSet = append(self.PSet, *temp)
	}
	return offspringSet
}

// 生成新一代popsize个解
func (self *NSGA_3) UpdateEXA(particleSet []basic_class.BasicSolution) {
	for i := 0; i < len(particleSet); i++ {
		domF := false
		j := 0
		for {
			if j >= len(self.Exa) {
				break
			}
			if self.BasicMooFunc.ParetoDominatesMin(particleSet[i].Objective, self.Exa[j].Objective) {
				self.Exa = append(self.Exa[:j], self.Exa[j+1:]...)
				j--
			} else if self.BasicMooFunc.ParetoDominatesMin(self.Exa[j].Objective, particleSet[i].Objective) {
				domF = true
				break
			}
			j++
		}
		// 如果新解particleSet[i]不被支配，则添加进EXA
		if domF == false {
			self.Exa = append(self.Exa, particleSet[i])
		}
	}

	// 如果EXA长度大于所需的个数
	if len(self.Exa) > self.PopSize {
		// 计算各个解的适应度
		self.CalPopFit(self.Exa)

		// 排序获得最差的EXA长度-PopSize个解
		for i := 0; i < len(self.Exa)-self.PopSize; i++ {
			for j := i; j < len(self.Exa); j++ {
				if self.Exa[i].TotalFit > self.Exa[j].TotalFit {
					temp := new(basic_class.BasicSolution)
					basic_class.Copy(self.Exa[j], temp)
					basic_class.Copy(self.Exa[i], &self.Exa[j])
					basic_class.Copy(*temp, &self.Exa[i])
				}
			}
		}
		self.Exa = self.Exa[len(self.Exa)-self.PopSize:]
	}
	// 找出最优解集
	front := self.FindFrontMinNoCon(self.Exa)
	// 更新GBest
	selIndex := util.RandomNumber(0, len(front)-1)
	basic_class.Copy(self.Exa[selIndex], &self.GBest)
}

// 两点交叉、变异算法
func (self *NSGA_3) GenNewPopBX(particleSet []basic_class.BasicSolution) []basic_class.BasicSolution {
	var offspringSet []basic_class.BasicSolution
	i := 0
	for {
		if len(offspringSet) > self.PopSize {
			break
		}
		var k, l int
		for {
			k = util.RandomNumber(0, len(particleSet)-1)
			if k != i {
				break
			}
		}
		for {
			l = util.RandomNumber(0, len(particleSet)-1)
			if l != k && l != i {
				break
			}
		}

		offspring := new(basic_class.BasicSolution)
		parent1 := particleSet[k]
		parent2 := particleSet[l]
		offspring.V = make([]float64, len(parent1.V))
		offspring.X = make([]float64, len(parent1.X))
		offspring.Solution = make([]int, len(parent1.Solution))

		index := 0
		for p := 0; p < basic_class.ProcessNum; p++ {
			for j := 0; j < basic_class.TaskNumPro; j++ {
				if Randomfloat64() < 0.5 {
					offspring.V[index] = parent1.V[index]
					offspring.X[index] = parent1.X[index]
					offspring.Solution[index] = parent1.Solution[index]
				} else {
					offspring.V[index] = parent2.V[index]
					offspring.X[index] = parent2.X[index]
					offspring.Solution[index] = parent2.Solution[index]
				}
				index++
			}
		}

		// 变异
		if Randomfloat64() < 0.15 {
			m1 := util.RandomNumber(0, self.TotalFunc-1) // 变异点

			nx := util.RandomNumber(0, len(self.CanSerAPro[m1])-1)
			offspring.Solution[m1] = self.CanSerAPro[m1][nx]
		}

		// 计算适应度
		offspring.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(offspring.Solution, self.CorFlag, self.PenFlag)

		// 更新最近点和最远点
		self.UpdateReference(*offspring)
		i++
		offspringSet = append(offspringSet, *offspring)
	}
	return offspringSet
}

// 返回帕累托前沿，目标值越小越好，不考虑约束
func (self *NSGA_3) FindFrontMinNoCon(inds []basic_class.BasicSolution) []basic_class.BasicSolution {
	var front []basic_class.BasicSolution
	// put the first guy in the front
	front = append(front, inds[0])

	// iterate over all the remaining individuals
	for i := 1; i < len(inds); i++ {
		ind := inds[i]
		noOneWasBetter := true

		// iterate over the entire front
		comfrontNum := 0
		for {
			if comfrontNum >= len(front) {
				break
			}
			frontmember := front[comfrontNum]

			// if the front member is better than the individual, dump the individual and go to the next one
			if self.BasicMooFunc.ParetoDominatesMin(frontmember.Objective, ind.Objective) {
				noOneWasBetter = false
				break
			} else if self.BasicMooFunc.ParetoDominatesMin(ind.Objective, frontmember.Objective) {
				front = append(front[:comfrontNum], front[comfrontNum+1:]...) // todo
			} else {
				comfrontNum++
			}
		}
		if noOneWasBetter {
			front = append(front, ind)
		}
	}
	return front
}

// 将连续值转化成有效的服务编号
func (self *NSGA_3) DouToInt1(value float64, index int) int {
	outValue := -1
	serCount := len(self.CanSerAPro[index])
	a := int(math.Abs(float64(int(value) % serCount)))
	outValue = self.CanSerAPro[index][a]
	return outValue
}

func (self *NSGA_3) UpdateReference(indiv basic_class.BasicSolution) {
	for j := 0; j < len(indiv.Objective); j++ {
		if indiv.Objective[j] < self.IdealPoint[j] {
			self.IdealPoint[j] = indiv.Objective[j]
		}

		if indiv.Objective[j] > self.NarPoint[j] {
			self.NarPoint[j] = indiv.Objective[j]
		}
	}
}

// 计算各个解的适应度
func (self *NSGA_3) CalPopFit(particleSet []basic_class.BasicSolution) {
	// Step1:计算cd,cv,meanCd,meanCv
	var cd []float64 // 存储收敛性cd
	var cv []float64 // 存储多样性cv
	cd = self.CalCd(particleSet)
	cv = self.CalCv(particleSet)
	meanCd := self.CalMean(particleSet, cd)
	meanCv := self.CalMean(particleSet, cv)

	// Step2:计算d1,d2,meanD1,meanD2
	var d1 []float64 // 存储收敛性d1
	var d2 []float64 // 存储多样性d2
	d1 = self.CalD1(particleSet)
	d2 = self.CalD2(particleSet, d1)
	meanD1 := self.CalMean(particleSet, d1)
	meanD2 := self.CalMean(particleSet, d2)

	// Step3:根据meanCd,meanCv,meanD1,meanD2确定各解的fitness
	self.CalIndFitness(particleSet, cd, cv, meanCd, meanCv, d1, d2, meanD1, meanD2)
}

func (self *NSGA_3) CalCd(particleSet []basic_class.BasicSolution) []float64 {
	popNum := len(particleSet)
	cd := make([]float64, popNum)  // 存储收敛性cd
	sde := make([]float64, popNum) // 存储每个解的收敛性度量
	for i := 0; i < popNum; i++ {
		cd = append(cd, 0.0)
		sde = append(cd, 0.0)
	}
	// Step1:计算sde[i] 公式（4）
	tMinSde := 100000000000.0
	tMaxSde := 0.0
	for i := 0; i < popNum; i++ {
		minSde := 1000000000.0
		for j := 0; j < popNum; j++ {
			if i != j {
				var sdeij float64
				for objNum := 0; objNum < basic_class.NrObj; objNum++ {
					delF := particleSet[j].Objective[objNum] - particleSet[i].Objective[objNum]
					if delF > 0 {
						sdeij += math.Pow(delF, 2)
					}
				}
				if sdeij < minSde {
					minSde = sdeij
				}
			}
		}
		sde[i] = minSde
		if minSde < tMinSde {
			tMinSde = minSde
		} else if minSde > tMaxSde {
			tMaxSde = minSde
		}
	}

	// Step2:归一计算Cd(pi)
	for i := 0; i < popNum; i++ {
		cd[i] = (sde[i] - tMinSde) / (tMaxSde - tMinSde)
	}
	return cd
}

// 计算多样性cv[i] 公式（7）、（6）//（6）计算Cv收敛性距离（7）计算dis欧式距离
func (self *NSGA_3) CalCv(particleSet []basic_class.BasicSolution) []float64 {
	popNum := len(particleSet)
	cv := make([]float64, popNum)
	dis := make([]float64, popNum)

	// Step1:计算dis 公式(7)
	for i := 0; i < popNum; i++ {
		a := 0.0
		for objNum := 0; objNum < basic_class.NrObj; objNum++ {
			a += math.Pow(particleSet[i].Objective[objNum]-self.IdealPoint[objNum], 2)
		}
		dis[i] = math.Sqrt(a)
	}

	// Step2:计算cv 公式(6)
	b := math.Sqrt(float64(basic_class.NrObj))
	for i := 0; i < popNum; i++ {
		cv[i] = 1 - dis[i]/b
	}
	return cv
}

// 计算meanCd、meanCv、meanD1、meanD2  //所有粒子平均值
func (self *NSGA_3) CalMean(particleSet []basic_class.BasicSolution, v []float64) float64 {
	sum := 0.0
	for i := 0; i < len(particleSet); i++ {
		sum += v[i]
	}
	mean := sum / float64(len(particleSet))
	return mean
}

// 计算d1[i] 公式（8）投影距离
func (self *NSGA_3) CalD1(particleSet []basic_class.BasicSolution) []float64 {
	popNum := len(particleSet)
	d1 := make([]float64, popNum) // 存储d1

	lenv := 0.0
	for objNum := 0; objNum < basic_class.NrObj; objNum++ {
		lenv += math.Pow(self.NarPoint[objNum]-self.IdealPoint[objNum], 2)
	}

	for i := 0; i < popNum; i++ {
		mul := 0.0
		for objNum := 0; objNum < basic_class.NrObj; objNum++ {
			mul += (particleSet[i].Objective[objNum] - self.IdealPoint[objNum]) * (self.NarPoint[objNum] - self.IdealPoint[objNum])
		}
		d1[i] = mul / math.Sqrt(lenv)
	}
	return d1
}

// 计算d2[i] 公式（9）垂直距离
func (self *NSGA_3) CalD2(particleSet []basic_class.BasicSolution, d1 []float64) []float64 {
	popNum := len(particleSet)
	d2 := make([]float64, popNum)

	for i := 0; i < popNum; i++ {
		mul := 0.0
		for objNum := 0; objNum < basic_class.NrObj; objNum++ {
			mul += math.Pow(particleSet[i].Objective[objNum]-self.IdealPoint[objNum], 2)
		}
		d2[i] = math.Sqrt(mul - math.Pow(d1[i], 2))
	}
	return d2
}

func (self *NSGA_3) CalIndFitness(particleSet []basic_class.BasicSolution, cd []float64, cv []float64, meanCd float64, meanCv float64, d1 []float64, d2 []float64, meanD1 float64, meanD2 float64) {
	for i := 0; i < len(particleSet); i++ {
		// Step1:确定公式（1）中的belta、alfa系数
		belta := 0.0
		alfa := 0.0
		if cv[i] > meanCv { // Case1: 原文为cv[i] < meanCv 梁改2020.3.12
			if d1[i] < meanD1 { // Case1.1
				belta = 1.0
				if cd[i] < meanCd { // Case1.1.1
					alfa = float64(util.RandomNumber(6, 12)) * 0.1
				} else { // Case1.1.2  cd[i] >= meanCd
					alfa = 1.0
				}
			} else { // Case1.2: d1[i] >= meanD1
				belta = 0.9
				if cd[i] < meanCd { // Case1.2.1
					alfa = 0.6
				} else { // Case1.2.2  cd[i] >= meanCd
					alfa = 0.9
				}
			}
		} else { // Case2: 原文为cv[i] >meanCv 改为 cv[i] <meanCv
			if d1[i] < meanD1 && d2[i] >= meanD2 { // Case2.1
				if cd[i] < meanCd { // Case2.1.1
					alfa = float64(util.RandomNumber(6, 12)) * 0.1
					belta = float64(util.RandomNumber(6, 12)) * 0.1
				} else { // Case2.1.2  cd[i] >= meanCd
					alfa = 1.0
					belta = 1.0
				}
			} else if d1[i] >= meanD1 || d2[i] < meanD2 { // Case2.2
				alfa = 0.2
				belta = 0.2
			}
		}
		// Step2:计算fitness
		particle := &particleSet[i]
		particle.TotalFit = alfa*cd[i] + belta*cv[i]
		if particleSet[i].TotalFit < 0.0001 {
			//xx := 0
		}
	}
}

// 单点交叉
func (self *NSGA_3) GenNewPop_1(i int) basic_class.BasicSolution {
	var j int
	for {
		j = self.NeighbourTable[i][util.RandomNumber(0, self.NeighbourSize-1)]
		if j != i {
			break
		}
	}
	basicPathdef1 := self.MainPop[j]
	basicPathdef2 := self.MainPop[i]
	var tempPath basic_class.BasicSolution
	tempPath.Solution = make([]int, basic_class.ProcessNum*basic_class.TaskNumPro)
	tempPath.Objective = []float64{}

	// 交叉
	cPoint := util.RandomNumber(0, self.TotalFunc-1)
	for k := 0; k < cPoint; k++ {
		tempPath.Solution[k] = basicPathdef1.Solution[k]
	}
	for k := cPoint; k < self.TotalFunc; k++ {
		tempPath.Solution[k] = basicPathdef2.Solution[k]
	}

	// 变异

	rand.Seed(time.Now().UnixNano())
	rs := rand.Float64()
	if rs < 0.05 {
		workNum := 0
		for p := 0; p < basic_class.ProcessNum; p++ { // 转换成适应度，计算适应度总和 //流程数量
			for a := 0; a < basic_class.TaskNumPro; a++ { // 该流程总活动数量
				nextCust := util.RandomNumber(0, len(self.CanSerAPro[workNum])-1)
				tempPath.Solution[workNum] = self.CanSerAPro[workNum][nextCust]
				workNum++
			}
		}
	} else if rs < 0.15 {
		m1 := util.RandomNumber(0, self.TotalFunc-1) // 变异点
		nx := util.RandomNumber(0, len(self.CanSerAPro[m1])-1)
		tempPath.Solution[m1] = self.CanSerAPro[m1][nx]
	}
	tempPath.Objective = self.ConstraintsFitness.CalFitnessMooNormalized(tempPath.Solution, self.CorFlag, self.PenFlag)
	return tempPath
}

//根据生成的offSpring,更新自己及邻居
//依据：当offSpring的PBI小于mainpop[weightindex]的PBI时，更新mainpop[weightindex]
func (self *NSGA_3) UpdateNeighbours_1(offspring basic_class.BasicSolution) {
	t := 0
	j := 0
	cnd := 0
	for {
		if t >= self.NeighbourSize {
			break
		}
		var weightIndex int
		if t == 0 {
			weightIndex = offspring.SubProbNo
		} else {
			j = util.RandomNumber(0, self.NeighbourSize-2)
			weightIndex = self.NeighbourTable[offspring.SubProbNo][j+1]
		}

		sol := self.MainPop[weightIndex]
		//if offspring.TchVal <= 0 || sol.TchVal <= 0 {
		//	xxx := 0
		//}
		if offspring.TchVal < sol.TchVal {
			offspring.CopyTo(&self.MainPop[weightIndex]) // 将offSpring覆盖mainpop[weightindex]，从而替代邻域中的一个解，
			self.MainPop[weightIndex].Trail = 0          // trail用于记录某个权重weightindex尝试被覆盖的次数
			cnd++
		} else {
			self.MainPop[weightIndex].Trail++
		}
		if cnd >= self.Nr { // nr=2表示子解最多可替换2个邻居解
			break
		}
	}
}

// 更新外部EXA，EXA不是存储的最优解，其存储的是特定数量的好解，每当新解来时，把密度最高的劣解去掉
func (self *NSGA_3) UpdateArchive(indiv basic_class.BasicSolution) {
	//Step1：计算indiv属于哪个权重向量
	//location := indiv.SubProbNo // location表示indiv的所属区域（基于角度）
}
