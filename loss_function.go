package decisiontrees

import (
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"math"
	"sort"
)

type LossFunction interface {
	GetGradient(e *Example) float64
	GetPrior(e Examples) float64
	GetLeafWeight(e Examples) float64
}

type LogitLoss struct {
	evaluator Evaluator
}

func (l LogitLoss) GetGradient(e *Example) float64 {
	prediction := l.evaluator.evaluate(e.Features)
	return 2 * e.Label / (1 + math.Exp(2*e.Label*prediction))
}

func (l LogitLoss) GetPrior(e Examples) float64 {
	sumLabels := float64(0.0)
	for _, example := range e {
		sumLabels += example.Label
	}
	averageLabel := sumLabels / float64(len(e))
	return 0.5 * math.Log((1+averageLabel)/(1-averageLabel))
}

func (l LogitLoss) GetLeafWeight(e Examples) float64 {
	numerator, denominator := 0.0, 0.0
	for _, example := range e {
		numerator += example.Label
		denominator += math.Abs(example.Label) * (2 - math.Abs(example.Label))
	}
	return numerator / denominator
}

type LeastAbsoluteDeviationLoss struct {
	evaluator Evaluator
}

func (l LeastAbsoluteDeviationLoss) GetPrior(e Examples) float64 {
	// Return the median label
	sort.Sort(LabelSorter{e})
	return e[e.Len()/2].Label
}

func (l LeastAbsoluteDeviationLoss) GetLeafWeight(e Examples) float64 {
	residuals := sort.Float64Slice(make([]float64, e.Len()))
	for i, _ := range e {
		residuals[i] = e[i].Label - l.evaluator.evaluate(e[i].Features)
	}
	residuals.Sort()
	return residuals[e.Len()/2]
}

func (l LeastAbsoluteDeviationLoss) GetGradient(e *Example) float64 {
	prediction := l.evaluator.evaluate(e.Features)
	if e.Label-prediction > 0 {
		return 1.0
	} else {
		return -1.0
	}
}

func NewLossFunction(l pb.LossFunction, evaluator Evaluator) LossFunction {
	switch l {
	case pb.LossFunction_LOGIT:
		return LogitLoss{
			evaluator: evaluator,
		}
	case pb.LossFunction_LEAST_ABSOLUTE_DEVIATION:
		return LeastAbsoluteDeviationLoss{
			evaluator: evaluator,
		}
	}
	panic(fmt.Sprint("Unknown enum: ", l.String()))
}
