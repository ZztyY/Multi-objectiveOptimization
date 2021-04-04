package MOEAs

//type NSGA3 struct {
//	NumObjectives int
//	Div           int
//	PopSize       int
//	NeighbourSize int
//	Weights       [][]float64
//	MainPop       []basic_class.MoChromosome
//	Mop           MultiObjectiveSolver
//	IgdValue      []float64
//	PofData       [][]float64
//	PofPath       string
//	Frm           plotFrm
//	ItrCounter    int
//	// todo extend Multi
//	IdealPoint []float64
//	NarPoint   []float64
//}
//
//func (self *NSGA3) Initial() {
//	self.IdealPoint = make([]float64, self.NumObjectives)
//	self.NarPoint = make([]float64, self.NumObjectives)
//	self.ItrCounter = 1
//
//	for i := 0; i < self.NumObjectives; i++ {
//		self.IdealPoint[i] = 1000000 // double max value
//		self.NarPoint[i] = 0         // double min value
//	}
//
//	self.InitWeight(self.Div)
//	self.InitialPopulation()
//	self.InitNeighbour()
//}
//
//func (self *NSGA3) InitNeighbour() {
//	var neighbourTable []int
//	distanceMatrix := make([][]float64, self.PopSize)
//	for i, _ := range distanceMatrix {
//		distanceMatrix[i] = make([]float64, self.PopSize)
//	}
//	for i := 0; i < self.PopSize; i++ {
//		distanceMatrix[i][i] = 0
//		for j := i + 1; j < self.PopSize; j++ {
//			distanceMatrix[i][j] = Distance(Weights[i], Weights[j])
//			distanceMatrix[j][i] = distanceMatrix[i][j]
//		}
//	}
//	for i := 0; i < self.PopSize; i++ {
//		val := make([]float64, self.PopSize)
//		for j := 0; j < self.PopSize; j++ {
//			val[j] = distanceMatrix[i][j]
//		}
//
//		index := Sort(val)
//		array := make([]int, self.NeighbourSize)
//		Copy(index, array, self.NeighbourSize)
//		for _, v := range array {
//			neighbourTable = append(neighbourTable, v)
//		}
//	}
//}
//
//func (self *NSGA3) InitWeight(m int) {
//	if self.NumObjectives < 6 {
//		self.Weights = UniPointGenerator.GetMUniDistributedPoint(self.NumObjectives, m)
//	} else {
//		self.Weights = UniPointGenerator.GetMaUniDistributedPoint(self.NumObjectives, m)
//	}
//	self.PopSize = len(self.Weights)
//
//}
//
//func (self *NSGA3) InitPopulation() {
//	for i := 0; i < self.PopSize; i++ {
//		chromosome := CreateChromosome()
//
//		Evaluate(chromosome)
//		UpdateReference(chromosome)
//		self.MainPop = append(self.MainPop, chromosome)
//	}
//}
//
//func (self *NSGA3) DoSolve() {
//	self.Initial()
//
//	prob := self.Mop.GetName()
//	if strings.Contains(prob, "DTLZ") {
//		self.IgdValue = append(self.IgdValue, QulityIndicator.QulityIndicator.DTLZIGD(self.MainPop, prob, self.NumObjectives))
//	} else {
//		file, _ = os.Open(self.PofPath + prob)
//		// self.PofData, _ =
//		self.IgdValue = append(self.IgdValue, QulityIndicator.QulityIndicator.IGD(self.MainPop, self.PofData))
//	}
//
//	//self.Frm =  todo
//	//self.Frm.Show()
//	//self.Frm.Refresh()
//	for {
//		if Terminated() {
//			break
//		}
//		offsPop := []basic_class.MoChromosome
//
//		for i := 0; i < self.PopSize; i++ {
//			var offSpring basic_class.MoChromosome
//			offspring = SBXCrossover(i, true) // GeneticOPDE//GeneticOPSBXCrossover
//			Evaluate(offspring)
//			offsPop = append(offsPop, offSpring)
//			UpdateReference(offspring)
//		}
//
//		var Pop []basic_class.MoChromosome
//		for _, v := range self.MainPop {
//			Pop = append(Pop, v)
//		}
//		for _, v := range offsPop {
//			Pop = append(Pop, v)
//		}
//
//		self.EnviromentSelection(Pop)
//
//		if self.ItrCounter % 10 == 0 {
//			self.Frm.RefreshPlot(self.ItrCounter, self.MainPop)
//			self.Frm.Refresh()
//
//			if strings.Contains(prob, "DTLZ") {
//				self.IgdValue = append(self.IgdValue, QulityIndicator.QulityIndicator.DTLZIGD(self.MainPop, prob, self.NumObjectives))
//			} else {
//				self.IgdValue = append(QulityIndicator.QulityIndicator.IGD(self.MainPop, self.PofData))
//			}
//		}
//
//		self.ItrCounter++
//	}
//	// todo write to file
//}
//
//func (self *NSGA3) EnviromentSelection(pop []basic_class.MoChromosome) {
//	result := []basic_class.MoChromosome
//	dominatedSet0 := NSGA.FastNonDominatedSort(pop)
//
//	if GlobalValue.IsNormalization {
//		UpdateNadirPoint(dominatedSet0[0])
//	}
//	//updateNadirPoint(dominatedSet0[0]);
//
//	cnt := 0
//	for {
//		if len(result) +len(dominatedSet0[cnt]) > self.PopSize {
//			break
//		}
//		for r := 0; r < len(dominatedSet0[cnt]); r++ {
//			dominatedSet0[cnt][r].Selected = true
//		}
//
//		for _, v := range dominatedSet0[cnt] {
//			result = append(result, v)
//		}
//		cnt++
//	}
//	if len(result) == self.PopSize {
//		self.MainPop.Clear()
//		for _, v := range result {
//			self.MainPop = append(self.MainPop, v)
//		}
//		return
//	}
//
//	count := make([]int, self.PopSize)
//	for i := 0; i < self.PopSize; i++ {
//		count[i] = 0
//	}
//
//	for i := 0; i < len(result); i++ {
//		// todo dist
//		pos := -1
//		for  j := 0; j < len(self.Weights); j++ {
//			dt := GetAngle(j, result[i], GlobalValue.IsNormalization)
//			if dt < dist {
//				dist = dt
//				pos = j
//			}
//		}
//		count[pos]++
//	}
//
//	var associatedSolution [][]MoChromosome
//	for i := 0; i < len(self.Weights); i++ {
//		var temp []MoChromosome
//		associatedSolution = append(associatedSolution, temp)
//	}
//
//	for i := 0; i < len(dominatedSet0[cnt]); i++ {
//		// todo dist
//		pos := -1
//		for j := 0; j < len(self.Weights); j++ {
//			dt := GetAngle(j, dominatedSet0[cnt][i], GlobalValue.IsNormalization)
//			if dt < dist {
//				dist = dt
//				pos = j
//			}
//		}
//		dominatedSet0[cnt][i].TchVal = PbiScalarObj(pos, dominatedSet0[cnt][i], GlobalValue.IsNormalization)
//		dominatedSet0[cnt][i].subProbNo = pos
//		associatedSolution[pos] = append(associatedSolution[pos], dominatedSet0[cnt][i])
//	}
//
//	for i := 0; i < len(self.Weights); i++ {
//		// todo associatedSolution[i].Sort()
//	}
//
//	var itr []int
//	for {
//		if len(result) >= self.PopSize {
//			break
//		}
//		itr.Clear()
//		min := self.PopSize
//		for i := 0; i < self.PopSize; i++ {
//			if min > count[i] && len(associatedSolution[i]) > 0 {
//				itr.Clear()
//				min = count[i]
//				itr = append(itr, i)
//			} else if min == count[i] && len(associatedSolution[i]) > 0 {
//				itr = append(itr, i)
//			}
//		}
//
//		pos := 1  // todo random
// 		count[itr[pos]]++
// 		result = append(result, associatedSolution[itr[pos]][0])
// 		associatedSolution[itr[pos]] = associatedSolution[itr[pos]][1:]
//	}
//	self.MainPop.Clear()
//	for _, v := range result {
//		self.MainPop = append(self.MainPop, v)
//	}
//	return
//}
