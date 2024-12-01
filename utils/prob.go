package utils

import (
	"math"
	"math/rand"
)

type ProbGenerator struct {
	probTransition func(float64) float64
	prob           float64
	probInitial    float64
}

func NewProbGenerator(avgProb float64) *ProbGenerator {
	probInitial := findProb(avgProb)
	println("Prob Generator: avgProb: ", avgProb, " probInitial: ", probInitial)
	return &ProbGenerator{
		probTransition: func(prob float64) float64 {
			return prob + probInitial
		},
		prob:        probInitial,
		probInitial: probInitial,
	}
}

func findProb(prob float64) float64 {
	left, right := 0.0, 1.0
	for right-left > 1e-9 {
		mid := (left + right) / 2
		p := calcExp(mid)
		if p < prob {
			left = mid
		} else {
			right = mid
		}
	}
	return left
}

func calcExp(probInitial float64) float64 {
	maxN := int(math.Ceil(1 / probInitial))

	exp := 0.0
	successProb := 0.0
	currentProb := 0.0

	for n := 1; n <= maxN; n++ {
		currentProb = (1 - successProb) * math.Min(1.0, float64(n)*probInitial)
		successProb += currentProb
		exp += float64(n) * currentProb
	}

	return 1 / exp
}

func (pg *ProbGenerator) Get() bool {
	num := rand.Float64()
	if num < pg.prob {
		pg.prob = pg.probInitial
		return true
	}
	pg.prob = pg.probTransition(pg.prob)
	return false
}

func (pg *ProbGenerator) GetInitialProb() float64 {
	return pg.probInitial
}
