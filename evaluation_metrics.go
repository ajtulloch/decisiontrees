package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"math"
	"sort"
)

type labelledPrediction struct {
	Label      bool
	Prediction float64
}

type labelledPredictions []labelledPrediction

func (l labelledPredictions) Len() int {
	return len(l)
}

func (l labelledPredictions) Swap(i int, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l labelledPredictions) Less(i int, j int) bool {
	return l[i].Prediction < l[j].Prediction
}

func (l labelledPredictions) ROC() float64 {
	sort.Sort(l)
	numPositives, numNegatives, weightedSum := 0, 0, 0
	for _, e := range l {
		if e.Label {
			numPositives += 1
		} else {
			numNegatives += 1
			weightedSum += numPositives
		}
	}
	return float64(weightedSum) / float64(numPositives*numNegatives)
}

func (l labelledPredictions) String() string {
	return fmt.Sprintf(
		"Size: %v\nROC: %v\nCalibration: %v\nNormalized Entropy: %v\nPositives: %v",
		l.Len(),
		l.ROC(),
		l.Calibration(),
		l.NormalizedEntropy(),
		l.numPositives())
}

func (l labelledPredictions) numPositives() int {
	s := 0
	for _, e := range l {
		if e.Label {
			s += 1
		}
	}
	return s
}

func (l labelledPredictions) LogScore() float64 {
	cumulativeLogLoss := 0.0
	for _, e := range l {
		if e.Label {
			cumulativeLogLoss += math.Log2(e.Prediction)
		} else {
			cumulativeLogLoss += math.Log2(1 - e.Prediction)
		}
	}
	return cumulativeLogLoss / float64(l.Len())
}

func (l labelledPredictions) Calibration() float64 {
	numPositives, sumPredictions := 0, 0.0
	for _, e := range l {
		sumPredictions += e.Prediction
		if e.Label {
			numPositives += 1
		}
	}
	return float64(sumPredictions) / float64(numPositives)
}

func (l labelledPredictions) NormalizedEntropy() float64 {
	numPositives := 0
	for _, e := range l {
		if e.Label {
			numPositives += 1
		}
	}
	p := float64(numPositives) / float64(l.Len())
	return l.LogScore() / (p*math.Log2(p) + (1-p)*math.Log2(1-p))
}

func computeEpochResult(e Evaluator, examples Examples) pb.EpochResult {
	l := make([]labelledPrediction, 0, len(examples))

	boolLabel := func(example *pb.Example) bool {
		if example.GetLabel() > 0 {
			return true
		}
		return false
	}

	for _, ex := range examples {
		l = append(l, labelledPrediction{
			Label:      boolLabel(ex),
			Prediction: e.Evaluate(ex.GetFeatures()),
		})
	}

	lp := labelledPredictions(l)
	return pb.EpochResult{
		Roc:               proto.Float64(lp.ROC()),
		LogScore:          proto.Float64(lp.LogScore()),
		NormalizedEntropy: proto.Float64(lp.NormalizedEntropy()),
		Calibration:       proto.Float64(lp.Calibration()),
	}
}

// LearningCurve computes the progressive learning curve after each epoch on the
// given examples
func LearningCurve(f *pb.Forest, e Examples) *pb.TrainingResults {
	tr := &pb.TrainingResults{
		EpochResults: make([]*pb.EpochResult, 0, len(f.GetTrees())),
	}

	for i := range f.GetTrees() {
		evaluator, err := NewRescaledFastForestEvaluator(&pb.Forest{
			Trees:     f.GetTrees()[:i],
			Rescaling: f.GetRescaling().Enum(),
		})
		if err != nil {
			glog.Fatal(err)
		}
		er := computeEpochResult(evaluator, e)
		tr.EpochResults = append(tr.EpochResults, &er)
	}
	return tr
}
