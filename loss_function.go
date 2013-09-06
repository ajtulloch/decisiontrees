package decisiontrees

import (
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"math"
)

type LossFunction interface {
	GetGradient(label float64, prediction float64) float64
	GetPrior(e Examples) float64
	GetLeafWeight(e Examples) float64
}

type L2LossFunction struct{}

func (l L2LossFunction) GetGradient(label float64, prediction float64) float64 {
	return 2 * (label - prediction)
}

type LogitLoss struct{}

func (l LogitLoss) GetGradient(label float64, prediction float64) float64 {
	return 2 * label / (1 + math.Exp(2*label*prediction))
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

type LeastAbsoluteDeviationLoss struct{}

func (l LeastAbsoluteDeviationLoss) GetGradient(label float64, prediction float64) float64 {
	if label-prediction > 0 {
		return 1.0
	} else {
		return -1.0
	}
}

func NewLossFunction(l pb.LossFunction) LossFunction {
	switch l {
	case pb.LossFunction_LOGIT:
		return LogitLoss{}
	}
	panic(fmt.Sprint("Unknown enum: ", l.String()))
}
