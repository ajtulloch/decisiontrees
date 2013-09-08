package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
)

func ConstructBoostingTree(e Examples, f pb.ForestConfig) *pb.Forest {
	forest := &pb.Forest{
		Trees: make([]*pb.TreeNode, 0, f.GetNumWeakLearners()),
	}

	lossFunction := NewLossFunction(
		f.GetLossFunctionConfig(),
		NewFastForestEvaluator(forest))

	// Initial prior
	forest.Trees = append(forest.Trees, &pb.TreeNode{
		LeafValue: proto.Float64(lossFunction.GetPrior(e)),
	})

	for i := 1; i < int(f.GetNumWeakLearners()); i++ {
		lossFunction := NewLossFunction(
			f.GetLossFunctionConfig(),
			NewFastForestEvaluator(forest))
		lossFunction.UpdateWeightedLabels(e)
		weakLearner := (&RegressionSplitter{}).GenerateTree(e)
		forest.Trees = append(forest.Trees, weakLearner)
	}

	return forest
}
