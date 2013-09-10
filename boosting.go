package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
)

type BoostingTreeGenerator struct {
	forestConfig *pb.ForestConfig
	forest       *pb.Forest
}

func (b *BoostingTreeGenerator) doInfluenceTrimming(e Examples) Examples {
	lossFunction := NewLossFunction(
		b.forestConfig.GetLossFunctionConfig(),
		NewFastForestEvaluator(b.forest))

	By(func(e1, e2 *Example) bool {
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

func (b *BoostingTreeGenerator) updateExampleWeights(e Examples) {
	lossFunction := NewLossFunction(
		b.forestConfig.GetLossFunctionConfig(),
		NewFastForestEvaluator(b.forest))
	lossFunction.UpdateWeightedLabels(e)
}

func (b *BoostingTreeGenerator) constructWeakLearner(e Examples) {
	lossFunction := NewLossFunction(
		b.forestConfig.GetLossFunctionConfig(),
		NewFastForestEvaluator(b.forest))

	weakLearner := (&RegressionSplitter{
		lossFunction:         lossFunction,
		splittingConstraints: b.forestConfig.GetSplittingConstraints(),
		shrinkageConfig:      b.forestConfig.GetShrinkageConfig(),
	}).GenerateTree(e)

	b.forest.Trees = append(b.forest.Trees, weakLearner)
}

func (b *BoostingTreeGenerator) doBoostingRound(e Examples, round int) {
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
}

func (b *BoostingTreeGenerator) initializeForest(e Examples) {
	b.forest = &pb.Forest{
		Trees: make([]*pb.TreeNode, 0, b.forestConfig.GetNumWeakLearners()),
	}

	lossFunction := NewLossFunction(
		b.forestConfig.GetLossFunctionConfig(),
		NewFastForestEvaluator(b.forest))

	// Initial prior
	b.forest.Trees = append(b.forest.Trees, &pb.TreeNode{
		LeafValue: proto.Float64(lossFunction.GetPrior(e)),
	})
}

func (b *BoostingTreeGenerator) ConstructBoostingTree(e Examples) *pb.Forest {
	b.initializeForest(e)
	for i := 1; i < int(b.forestConfig.GetNumWeakLearners()); i++ {
		b.doBoostingRound(e, i)
	}
	return b.forest
}
