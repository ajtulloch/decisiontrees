package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"log"
)

func ConstructBoostingTree(e Examples, f pb.ForestConfig) *pb.Forest {
	forest := &pb.Forest{
		Trees: make([]*pb.TreeNode, 0, f.GetNumWeakLearners()),
	}

	lossFunction := NewLossFunction(f.GetLossFunction())

	// Initial prior
	forest.Trees = append(forest.Trees, &pb.TreeNode{
		LeafValue: proto.Float64(lossFunction.GetPrior(e)),
	})

	weightedExamples := make([]*Example, len(e))
	numCopied := copy(weightedExamples, e)
	if numCopied != len(e) {
		log.Fatal("Failed copying all examples")
	}

	for i := 1; i < int(f.GetNumWeakLearners()); i++ {
		evaluator := NewFastForestEvaluator(forest)
		for i, example := range e {
			gradient := lossFunction.GetGradient(
				example.Label,
				evaluator.evaluate(example.Features))
			weightedExamples[i].Label = gradient
		}

		weakLearner := (&RegressionSplitter{}).GenerateTree(weightedExamples)
		forest.Trees = append(forest.Trees, weakLearner)
	}

	return forest
}
