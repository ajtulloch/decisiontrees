package decisiontrees

import (
	"math"
	"math/rand"
	"testing"
)

var dataset = LabelledPredictions([]LabelledPrediction{
	LabelledPrediction{false, 0.0},
	LabelledPrediction{true, 0.0},
})

func randomDataset(size int, average float64) LabelledPredictions {
	predictions := LabelledPredictions(make([]LabelledPrediction, size))
	for i, _ := range predictions {
		predictions[i].Prediction = average
		predictions[i].Label = rand.Float64() < average
	}
	return predictions
}
func near(expected float64, actual float64) bool {
	return math.Abs(expected-actual) < 0.02
}

func TestRandomDatasetHasExpectedStatistics(t *testing.T) {
	tests := []struct {
		numSamples  int
		probability float64
	}{
		{100000, 0.02},
		{100000, 0.5},
		{100000, 0.9},
	}

	for _, tt := range tests {
		d := randomDataset(tt.numSamples, tt.probability)
		t.Log()
		t.Log(d.String())
		if !near(d.Calibration(), 1.0) {
			t.Errorf("Calibration: expected %v, had %v", 1.0, d.Calibration())
		}

		expectedLogScore :=
			tt.probability*math.Log2(tt.probability) +
				(1-tt.probability)*math.Log2(1-tt.probability)
		if !near(d.LogScore(), expectedLogScore) {
			t.Errorf("Logscore: expected %v, had %v", expectedLogScore, d.LogScore())
		}

		if !near(d.NormalizedEntropy(), 1.0) {
			t.Errorf("Entropy: expected %v, had %v", 1.0, d.NormalizedEntropy())
		}

		if !near(d.ROC(), 0.5) {
			t.Errorf("ROC: expected %v, had %v", 0.5, d.ROC())
		}
	}
}
