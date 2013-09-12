package decisiontrees

import (
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"math"
)

type LossFunction interface {
	UpdateWeightedLabels(e Examples)
	GetPrior(e Examples) float64
	GetLeafWeight(e Examples) float64
	GetSampleImportance(ex *Example) float64
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

func (l LogitLoss) GetSampleImportance(ex *Example) float64 {
	prediction := l.evaluator.evaluate(ex.Features)
	weightedLabel := 2 * ex.Label / (1 + math.Exp(2*ex.Label*prediction))
	return math.Abs(weightedLabel) * (2 - math.Abs(weightedLabel))
}

func clampToRange(value, lower, upper float64) float64 {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}

const (
	minLogitPrior = -20.0
	maxLogitPrior = 20.0
)

func (l LogitLoss) GetPrior(e Examples) float64 {
	if len(e) == 0 {
		return 0.0
	}

	sumLabels := float64(0.0)
	for _, example := range e {
		sumLabels += example.Label
	}
	averageLabel := sumLabels / float64(len(e))
	return clampToRange(
		0.5*math.Log((1+averageLabel)/(1-averageLabel)),
		minLogitPrior,
		maxLogitPrior)
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

func (l LeastAbsoluteDeviationLoss) GetSampleImportance(ex *Example) float64 {
	return 1.0
}

func (l LeastAbsoluteDeviationLoss) GetPrior(e Examples) float64 {
	// Return the median label
	By(func(e1, e2 *Example) bool { return e1.Label < e2.Label }).Sort(e)
	return e[len(e)/2].Label
}

func (l LeastAbsoluteDeviationLoss) residual(ex *Example) float64 {
	return ex.Label - l.evaluator.evaluate(ex.Features)
}

func (l LeastAbsoluteDeviationLoss) GetLeafWeight(e Examples) float64 {
	By(func(e1, e2 *Example) bool {
		return l.residual(e1) < l.residual(e2)
	}).Sort(e)
	return l.residual(e[len(e)/2])
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
	By(func(e1, e2 *Example) bool { return e1.Label < e2.Label }).Sort(e)
	return e[len(e)/2].Label
}

func (h HuberLoss) GetSampleImportance(ex *Example) float64 {
	return 1.0
}

func (h HuberLoss) residual(ex *Example) float64 {
	return ex.Label - h.evaluator.evaluate(ex.Features)
}

func (h HuberLoss) UpdateWeightedLabels(e Examples) {
	By(func(e1, e2 *Example) bool {
		return h.residual(e1) < h.residual(e2)
	}).Sort(e)
	marginalExample := e[int64(float64(len(e))*h.huberAlpha)]
	delta := h.residual(marginalExample)
	for _, ex := range e {
		divergence := h.residual(ex)
		if divergence <= delta {
			ex.WeightedLabel = divergence
		} else {
			ex.WeightedLabel = delta * divergence / math.Abs(divergence)
		}
	}
}

func (h HuberLoss) GetLeafWeight(e Examples) float64 {
	By(func(e1, e2 *Example) bool {
		return h.residual(e1) < h.residual(e2)
	}).Sort(e)
	medianResidual := h.residual(e[len(e)/2])
	innerDistribution := 0.0
	for _, ex := range e {
		residualDelta := h.residual(ex) - medianResidual
		if residualDelta == 0.0 {
			continue
		}

		innerDistribution +=
			residualDelta / math.Abs(residualDelta) *
				math.Min(h.lastDeltaM, math.Abs(residualDelta))
	}

	return medianResidual + innerDistribution/float64(len(e))
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
	glog.Fatalf("Unknown enum: %v", l)
	panic("")
}
