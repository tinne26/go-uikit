package layout

import (
	"cmp"
	"math/rand/v2"
	"testing"
)

func TestRowStackPattern(t *testing.T) {
	var buff []int
	for i := range 500 {
		var weights []float64
		for range 1 + rand.IntN(16) {
			weights = append(weights, rand.Float64())
		}
		rsp := rowStackPattern{Gap: rand.IntN(56), Weights: weights}
		rsp.normalize()
		wsum := sum(rsp.Weights)
		if wsum > 1.0000001 || wsum < 0.9999999 {
			t.Fatalf("expected sum = 1.0, got %f", wsum)
		}
		totalWidth := int(float64(rsp.Gap*(len(weights)-1)) * (1.0 + rand.Float64()*6))
		buff = rsp.computeWidths(totalWidth, buff)
		actualTotal := sum(buff) + rsp.Gap*(len(weights)-1)
		if actualTotal != totalWidth {
			t.Fatalf("test#%d: expected %d, got %d | gap: %d, weights %v", i, totalWidth, actualTotal, rsp.Gap, rsp.Weights)
		}
	}
}

func sum[T cmp.Ordered](vs []T) T {
	var s T
	for _, v := range vs {
		s += v
	}
	return s
}
