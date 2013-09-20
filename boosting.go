package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"time"
)

type boostingTreeGenerator struct {
	forestConfig *pb.ForestConfig
	forest       *pb.Forest
}

func (b *boostingTreeGenerator) doInfluenceTrimming(e Examples) Examples {
	lossFunction := b.getLossFunction()

	by(func(e1, e2 *pb.Example) bool {
		return lossFunction.GetSampleImportance(e1) < lossFunction.GetSampleImportance(e2)
	}).Sort(e)

	// Find cutoff point
	weightSum := 0.0
	for _, ex := range e {
		weightSum += lossFunction.GetSampleImportance(ex)
	}

	cutoffPointSum := b.forestConfig.GetInfluenceTrimmingConfig().GetAlpha() * weightSum
	cutoffPoint, cumulativeSum := 0, 0.0
	for i, ex := range e {
		cutoffPoint = i
		if cumulativeSum < cutoffPointSum {
			break
		}
		cumulativeSum += lossFunction.GetSampleImportance(ex)
	}
	return e[cutoffPoint:]
}

func (b *boostingTreeGenerator) updateExampleWeights(e Examples) {
	b.getLossFunction().UpdateWeightedLabels(e)
}

func (b *boostingTreeGenerator) constructWeakLearner(e Examples) {
	weakLearner := (&regressionSplitter{
		lossFunction:         b.getLossFunction(),
		splittingConstraints: b.forestConfig.GetSplittingConstraints(),
		shrinkageConfig:      b.forestConfig.GetShrinkageConfig(),
	}).GenerateTree(e)

	b.forest.Trees = append(b.forest.Trees, weakLearner)
}

func (b *boostingTreeGenerator) doBoostingRound(e Examples, round int) {
	startTime := time.Now()
	defer func() {
		glog.Infof("Round %v, duration %v", round, time.Now().Sub(startTime))
	}()

	if b.forestConfig.GetStochasticityConfig() != nil {
		e = e.subsampleExamples(b.forestConfig.GetStochasticityConfig().GetPerRoundSamplingRate())
	}

	// Trim the low-sample influencers
	if b.forestConfig.GetInfluenceTrimmingConfig() != nil &&
		b.forestConfig.GetInfluenceTrimmingConfig().GetWarmupRounds() < int64(round) {
		e = b.doInfluenceTrimming(e)
	}

	b.updateExampleWeights(e)
	b.constructWeakLearner(e)

	metrics := b.computeTrainingMetrics(e)
	glog.Infof("Epoch: %v, Metrics: %+v", round, metrics)
}

func (b *boostingTreeGenerator) computeTrainingMetrics(e Examples) pb.EpochResult {
	evaluator, err := NewFastForestEvaluator(b.forest)
	if err != nil {
		glog.Fatal(err)
	}

	return computeEpochResult(evaluator, e)
}

func (b *boostingTreeGenerator) getLossFunction() LossFunction {
	evaluator, err := NewFastForestEvaluator(b.forest)
	if err != nil {
		glog.Fatal(err)
	}

	return NewLossFunction(b.forestConfig.GetLossFunctionConfig(), evaluator)
}

func (b *boostingTreeGenerator) initializeForest(e Examples) {
	b.forest = &pb.Forest{
		Trees: make([]*pb.TreeNode, 0, b.forestConfig.GetNumWeakLearners()),
	}

	// Initial prior
	b.forest.Trees = append(b.forest.Trees, &pb.TreeNode{
		LeafValue: proto.Float64(b.getLossFunction().GetPrior(e)),
	})
}

func (b *boostingTreeGenerator) ConstructBoostingTree(e Examples) *pb.Forest {
	glog.Infof("Initializing forest with config %+v", b.forestConfig)
	b.initializeForest(e)
	for i := 0; i < int(b.forestConfig.GetNumWeakLearners()); i++ {
		glog.Infof("Running boosting round %v", i)
		b.doBoostingRound(e, i)
	}
	return b.forest
}

// NewBoostingTreeGenerator returns a boosting tree given the forest config
func NewBoostingTreeGenerator(forestConfig *pb.ForestConfig) *boostingTreeGenerator {
	return &boostingTreeGenerator{
		forestConfig: forestConfig,
	}
}
