package algorithms

import (
	"Multi-objectiveOptimization/basic_class"
	"Multi-objectiveOptimization/util"
	"math"
	"math/rand"
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
}

// 初始化
func (self *MMPso) Init(popSize int, totalFunc int, parallelFunc int, generation int, w float64, c1 float64, c2 float64, c3 float64) {
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
}

func (self *MMPso) Run() {
	self.StartTime = time.Now()
	self.CanSerAPro = self.MoeaqiBt.ServicePreProcess()
	self.PSet = self.randomPathinitial()

	self.UpdateEXAIni(self.PSet) // 在初始化阶段，更新EXA为初始解集中的最优解PSet-pbest

	for q := 0; q < self.Generation/2; q++ {
		// 根据particleSet生成popsize个新解
		offspringSet := self.GenNewPopPSO(self.PSet) // 通过PSO算法对Pset初始解集操作生成后代offSpring

		// 筛选前popsize个新解，放入pop，子代best
		self.UpdateEXA(offspringSet)

		// 根据EXA生成popsize个新解
		SSet := self.GenNewPopBX(self.Exa) // 交叉变异

		// 筛选前popsize个新解，放入pop种群
		self.UpdateEXA(SSet) // 对新解更新

		self.EndTime = time.Now()
		// self.TimeSpan = self.EndTime - self.StartTime todo
		self.TimeConsume = self.TimeSpan.Milliseconds()

		// 将归一化的目标值重新计算为QoS值
		for i := 0; i < len(self.Exa); i++ {
			self.Exa[i].Objective = self.ConstraintsFitness.CalFitnessMoo(self.Exa[i].Solution, true, true)
			conCount := self.ConstraintsFitness.CalTotalConstraint(self.Exa[i].Solution, true, true, true)
			self.Exa[i].Objective = append(self.Exa[i].Objective, conCount)
		}

		self.ArrResult = self.Exa
	}
}

func (self *MMPso) randomPathinitial() []basic_class.BasicSolution {
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
func (self *MMPso) UpdateEXAIni(particleSet []basic_class.BasicSolution) {
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
func (self *MMPso) GenNewPopPSO(particleSet []basic_class.BasicSolution) []basic_class.BasicSolution {
	offspringSet := make([]basic_class.BasicSolution, self.PopSize)
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
func (self *MMPso) UpdateEXA(particleSet []basic_class.BasicSolution) {
	for i := 0; i < len(particleSet); i++ {
		domF := false
		j := 0
		for {
			if j >= len(self.Exa) {
				break
			}
			if self.BasicMooFunc.ParetoDominatesMin(particleSet[i].Objective, self.Exa[j].Objective) {
				self.Exa = append(self.Exa[:j], self.Exa[j:]...)
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
func (self *MMPso) GenNewPopBX(particleSet []basic_class.BasicSolution) []basic_class.BasicSolution {
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
func (self *MMPso) FindFrontMinNoCon(inds []basic_class.BasicSolution) []basic_class.BasicSolution {
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
				front = append(front[:comfrontNum], front[comfrontNum:]...)
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
func (self *MMPso) DouToInt1(value float64, index int) int {
	outValue := -1
	serCount := len(self.CanSerAPro[index])
	a := int(math.Abs(float64(int(value) % serCount)))
	outValue = self.CanSerAPro[index][a]
	return outValue
}

func (self *MMPso) UpdateReference(indiv basic_class.BasicSolution) {
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
func (self *MMPso) CalPopFit(particleSet []basic_class.BasicSolution) {
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

func (self *MMPso) CalCd(particleSet []basic_class.BasicSolution) []float64 {
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
func (self *MMPso) CalCv(particleSet []basic_class.BasicSolution) []float64 {
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
func (self *MMPso) CalMean(particleSet []basic_class.BasicSolution, v []float64) float64 {
	sum := 0.0
	for i := 0; i < len(particleSet); i++ {
		sum += v[i]
	}
	mean := sum / float64(len(particleSet))
	return mean
}

// 计算d1[i] 公式（8）投影距离
func (self *MMPso) CalD1(particleSet []basic_class.BasicSolution) []float64 {
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
func (self *MMPso) CalD2(particleSet []basic_class.BasicSolution, d1 []float64) []float64 {
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

func (self *MMPso) CalIndFitness(particleSet []basic_class.BasicSolution, cd []float64, cv []float64, meanCd float64, meanCv float64, d1 []float64, d2 []float64, meanD1 float64, meanD2 float64) {
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
