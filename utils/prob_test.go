package utils

import "testing"

func TestPb(t *testing.T) {
	avgProb := 0.1
	pg := NewProbGenerator(avgProb)
	t.Log("Initial prob: ", pg.GetInitialProb())
	for i := 0; i < 10; i++ {
		t.Log(pg.Get())
	}
}
