package decisiontrees

import (
	// "code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	// "github.com/golang/glog"
	"sync"
)

func averageLabel(e Examples) float64 {
	result := 0.0
	for _, ex := range e {
		result += ex.GetLabel()
	}
	return result / float64(len(e))
}

type randomForestGenerator struct {
	forestConfig *pb.ForestConfig
}

func (r *randomForestGenerator) constructRandomTree(e Examples) *pb.TreeNode {
	splitter := regressionSplitter{
		leafWeight: averageLabel,
		featureSelector: randomForestFeatureSelector{
			int(r.forestConfig.GetStochasticityConfig().GetFeatureSampleSize()),
		},
		splittingConstraints: r.forestConfig.GetSplittingConstraints(),
		shrinkageConfig:      r.forestConfig.GetShrinkageConfig(),
	}
	return splitter.GenerateTree(e.boostrapExamples(
		r.forestConfig.GetStochasticityConfig().GetExampleBoostrapProportion()))
}

func (r *randomForestGenerator) ConstructForest(e Examples) *pb.Forest {
	result := &pb.Forest{
		Trees:     make([]*pb.TreeNode, int(r.forestConfig.GetNumWeakLearners())),
		Rescaling: pb.Rescaling_AVERAGING.Enum(),
	}

	wg := sync.WaitGroup{}
	for i := 0; i < int(r.forestConfig.GetNumWeakLearners()); i++ {
		wg.Add(1)
		go func(i int) {
			result.Trees[i] = r.constructRandomTree(e)
		}(i)
	}
	return result
}
