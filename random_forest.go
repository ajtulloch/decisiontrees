package decisiontrees

import (
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"sync"
)

func ConstructRandomForest(e Examples, f pb.ForestConfig) *pb.Forest {
	forest := &pb.Forest{
		Trees: make([]*pb.TreeNode, 0, f.GetNumWeakLearners()),
	}

	w := sync.WaitGroup{}
	for i := 0; i < int(f.GetNumWeakLearners()); i++ {
		w.Add(1)
		go func(i int) {
			forest.Trees[i] = (&RegressionSplitter{}).GenerateTree(e.boostrapSample(100))
			w.Done()
		}(i)
	}
	w.Wait()
	return forest
}
