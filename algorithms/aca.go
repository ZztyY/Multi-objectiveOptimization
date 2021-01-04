package algorithms

import (
	"github.com/chfenger/goNum"
	"github.com/emirpasic/gods/sets/hashset"
	"gonum.org/v1/gonum/floats"
	"math/rand"
	"time"
)

type AcaTsp struct {
	Func               func(map[string]interface{}) float64 `json:"func"`
	NDim               int                                  `json:"n_dim"`                // 城市数量
	SizePop            int                                  `json:"size_pop"`             // 蚂蚁数量
	MaxIter            int                                  `json:"max_iter"`             // 迭代次数
	Alpha              int                                  `json:"alpha"`                // 信息素重要程度
	Beta               int                                  `json:"beta"`                 // 适应度的重要程度
	Rho                float64                              `json:"rho"`                  // 信息素挥发速度
	ProbMatrixDistance goNum.Matrix                         `json:"prob_matrix_distance"` // 避免除零错误
	Tau                goNum.Matrix                         `json:"tau"`                  // 信息素矩阵，每次迭代都会更新
	Table              goNum.Matrix                         `json:"table"`                // 某一代每个蚂蚁的爬行路径
	Y                  int                                  `json:"y"`                    // 某一代每个蚂蚁的爬行总距离
	GenerationBestX    [][]float64                          `json:"generation_best_x"`    // 记录各代的最佳情况
	GenerationBestY    []float64                            `json:"generation_best_y"`    // 记录各代的最佳情况
	XBestHistory       []interface{}                        `json:"x_best_history"`       // 历史
	YBestHistory       []interface{}                        `json:"y_best_history"`       // 历史
	BestX              []float64                            `json:"best_x"`
	BestY              float64                              `json:"best_y"`
}

func (self *AcaTsp) Init(nDim int, sizePop int, maxIter int, alpha int, beta int, rho float64, distanceMatrix goNum.Matrix, f func(map[string]interface{}) float64) {
	self.NDim = nDim       // 城市数量
	self.SizePop = sizePop // 蚂蚁数量
	self.MaxIter = maxIter // 迭代次数
	self.Alpha = alpha     // 信息素重要程度
	self.Beta = beta       // 适应度的重要程度
	self.Rho = rho         // 信息素挥发速度

	self.ProbMatrixDistance.Data = goNum.NumProductMatrix(goNum.IdentityE(self.NDim), 1e-10).Data // 避免除零错误
	self.ProbMatrixDistance.Rows = self.NDim
	self.ProbMatrixDistance.Columns = self.NDim
	for k, v := range self.ProbMatrixDistance.Data {
		v = v + distanceMatrix.Data[k]
		v = 1 / v
	}

	self.Tau = goNum.ZeroMatrix(self.NDim, self.NDim) // 信息素矩阵，每次迭代都会更新
	for i := 0; i < len(self.Tau.Data); i++ {
		self.Tau.Data[i] = 1.0
	}
	self.Table = goNum.ZeroMatrix(self.SizePop, self.NDim) // 某一代每个蚂蚁的爬行路径
	self.Y = 0                                             // 某一代每个蚂蚁的爬行总距离
	self.GenerationBestX = nil                             // 记录各代的最佳情况
	self.GenerationBestY = nil                             // 记录各代的最佳情况
	self.BestX = nil
	self.BestY = 0
	self.Func = f
}

func (self *AcaTsp) Run() ([]float64, float64) {
	// 每次迭代
	for i := 0; i < self.MaxIter; i++ {
		// 转移概率，无需归一化
		temp := goNum.IdentityE(self.NDim)
		for j := 0; j < self.Alpha; j++ {
			temp = goNum.DotPruduct(temp, self.Tau)
		}
		temp = goNum.DotPruduct(temp, self.ProbMatrixDistance)
		proMatrix := goNum.IdentityE(self.NDim)
		for j := 0; j < self.Beta; j++ {
			proMatrix = goNum.DotPruduct(proMatrix, temp)
		}

		// 每只蚂蚁
		for j := 0; j < self.SizePop; j++ {
			self.Table.SetMatrix(j, 0, 0) // start point 可以随机

			// 每个蚂蚁到达的每个节点
			for k := 0; k < self.NDim-1; k++ {
				rows := self.Table.RowOfMatrix(j)
				// 已经经过的点和当前点不能再次经过
				allowSet := hashset.New()
				for v := 0; v < self.NDim; v++ {
					allowSet.Add(v)
				}
				// 在这些点中选择
				for _, v := range rows[0 : k+1] {
					allowSet.Remove(int(v))
				}
				allowList := allowSet.Values()
				probTemp := proMatrix.RowOfMatrix(int(self.Table.GetFromMatrix(j, k)))
				var prob []float64
				for _, v := range allowList {
					prob = append(prob, probTemp[v.(int)])
				}
				sum := floats.Sum(prob)
				// 概率归一化
				for _, v := range prob {
					v = v / sum
				}
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				num := r.Float64()
				sum = 0
				randSelect := 0
				for m, v := range prob {
					sum += v
					if num <= sum {
						randSelect = m
						break
					}
				}
				nextPoint := allowList[randSelect]
				self.Table.SetMatrix(j, k+1, float64(nextPoint.(int)))
			}
		}
		// 计算距离
		var y []float64
		for j := 0; j < self.Table.Rows; j++ {
			m := map[string]interface{}{
				"n_dim":   self.NDim,
				"routine": self.Table.RowOfMatrix(j),
			}
			y = append(y, self.Func(m))
		}

		// 记录历史最好情况
		minimum, _, _ := goNum.Min(y)
		indexBest := 0
		for k, v := range y {
			if v == minimum {
				indexBest = k
				break
			}
		}
		xBest := self.Table.RowOfMatrix(indexBest)
		yBest := y[indexBest]
		self.GenerationBestX = append(self.GenerationBestX, xBest)
		self.GenerationBestY = append(self.GenerationBestY, yBest)

		// 计算需要重新涂抹的信息素
		deltaTau := goNum.ZeroMatrix(self.NDim, self.NDim)
		// 每只蚂蚁
		for j := 0; j < self.SizePop; j++ {
			// 每个节点
			for k := 0; k < self.NDim-1; k++ {
				n1 := int(self.Table.GetFromMatrix(j, k))
				n2 := int(self.Table.GetFromMatrix(k, k+1))
				deltaTau.SetMatrix(n1, n2, deltaTau.GetFromMatrix(n1, n2)+1/float64(y[j])) // 涂抹的信息素
			}
			n1 := int(self.Table.GetFromMatrix(j, self.NDim-1))
			n2 := int(self.Table.GetFromMatrix(j, 0))
			deltaTau.SetMatrix(n1, n2, deltaTau.GetFromMatrix(n1, n2)+1/float64(y[j]))
		}

		// 信息素飘散+信息素涂抹
		self.Tau = goNum.AddMatrix(goNum.NumProductMatrix(self.Tau, 1-self.Rho), deltaTau)
	}
	minimum, _, _ := goNum.Min(self.GenerationBestY)
	bestGeneration := 0
	for k, v := range self.GenerationBestY {
		if v == minimum {
			bestGeneration = k
		}
	}
	self.BestX = self.GenerationBestX[bestGeneration]
	self.BestY = self.GenerationBestY[bestGeneration]
	return self.BestX, self.BestY
}
