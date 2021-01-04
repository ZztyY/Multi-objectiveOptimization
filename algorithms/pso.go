package algorithms

import (
	"github.com/chfenger/goNum"
)

type Pso struct {
	NDim          int
	Pop           int
	MaxIter       int
	Lb            goNum.Matrix
	Ub            goNum.Matrix
	W             float64
	Cp            float64
	Cg            float64
	Verbose       bool
	HasConstraint bool
	ConstraintUeq []func([]int) int
}

func (self *Pso) Init() {

}

// gather all unequal constraint functions
func (self *Pso) Check_constraint(x []int) bool {
	for _, constraintFunc := range self.ConstraintUeq {
		if constraintFunc(x) > 0 {
			return false
		}
	}
	return true
}

func (self *Pso) UpdateV() {

}
