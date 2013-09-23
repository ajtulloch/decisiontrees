package decisiontrees

import (
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
)

// ForestGenerator is implemented by various algorithms that generate
// an ensemble of decision trees from the given training dataset.
type ForestGenerator interface {
	ConstructForest(e Examples) *pb.Forest
}

// NewForestGenerator returns a ForeestGenerator from the given
// ForestConfig.
func NewForestGenerator(forestConfig *pb.ForestConfig) (ForestGenerator, error) {
	switch forestConfig.GetAlgorithm() {
	case pb.Algorithm_BOOSTING:
		return &boostingTreeGenerator{forestConfig: forestConfig}, nil
	case pb.Algorithm_RANDOM_FOREST:
		return &randomForestGenerator{forestConfig: forestConfig}, nil
	}
	return nil, fmt.Errorf("unknown algorithm type: %v", forestConfig.GetAlgorithm())
}
