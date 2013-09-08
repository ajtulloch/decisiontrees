package decisiontrees

import (
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"math"
	"sort"
)

type LossFunction interface {
	UpdateWeightedLabels(e Examples)
	GetPrior(e Examples) float64
	GetLeafWeight(e Examples) float64
}

type LogitLoss struct {
	evaluator Evaluator
}

func (l LogitLoss) UpdateWeightedLabels(e Examples) {
	for _, ex := range e {
		prediction := l.evaluator.evaluate(ex.Features)
		ex.WeightedLabel = 2 * ex.Label / (1 + math.Exp(2*ex.Label*prediction))
	}
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

func (l LeastAbsoluteDeviationLoss) UpdateWeightedLabels(e Examples) {
	for _, ex := range e {
		prediction := l.evaluator.evaluate(ex.Features)
		if ex.Label-prediction > 0 {
			ex.WeightedLabel = 1.0
		} else {
			ex.WeightedLabel = -1.0
		}
	}
}

type HuberLoss struct {
	huberAlpha float64
	evaluator  Evaluator

	// Somewhat janky
	lastDeltaM float64
}

func (h HuberLoss) GetPrior(e Examples) float64 {
	// Return the median label
	sort.Sort(LabelSorter{e})
	return e[e.Len()/2].Label
}

func (h HuberLoss) UpdateWeightedLabels(e Examples) {
	residuals := sort.Float64Slice(make([]float64, e.Len()))
	for i, ex := range e {
		residuals[i] = ex.Label - h.evaluator.evaluate(ex.Features)
	}
	residuals.Sort()
	delta := residuals[int64(float64(e.Len())*h.huberAlpha)]
	for _, ex := range e {
		divergence := ex.Label - h.evaluator.evaluate(ex.Features)
		if divergence <= delta {
			ex.WeightedLabel = divergence
		} else {
			ex.WeightedLabel = delta * divergence / math.Abs(divergence)
		}
	}
}

func (h HuberLoss) GetLeafWeight(e Examples) float64 {
	residuals := sort.Float64Slice(make([]float64, e.Len()))
	for i, ex := range e {
		residuals[i] = ex.Label - h.evaluator.evaluate(ex.Features)
	}
	residuals.Sort()
	medianResidual := residuals[e.Len()/2]

	innerDistribution := 0.0
	for _, ex := range e {
		residualDelta :=
			(ex.Label - h.evaluator.evaluate(ex.Features)) -
				medianResidual

		if residualDelta == 0.0 {
			continue
		}

		innerDistribution +=
			residualDelta / math.Abs(residualDelta) *
				math.Min(h.lastDeltaM, math.Abs(residualDelta))
	}

	return medianResidual + innerDistribution/float64(e.Len())
}

func NewLossFunction(l *pb.LossFunctionConfig, evaluator Evaluator) LossFunction {
	switch l.GetLossFunction() {
	case pb.LossFunction_LOGIT:
		return LogitLoss{
			evaluator: evaluator,
		}
	case pb.LossFunction_LEAST_ABSOLUTE_DEVIATION:
		return LeastAbsoluteDeviationLoss{
			evaluator: evaluator,
		}
	case pb.LossFunction_HUBER:
		return HuberLoss{
			huberAlpha: l.GetHuberAlpha(),
			evaluator:  evaluator,
		}
	}
	panic(fmt.Sprint("Unknown enum: ", l.String()))
}
