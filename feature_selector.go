package decisiontrees

import (
	"math/rand"
)

// FeatureSelector allows algorithms to configure which
// features to use for a given round of splitting
type FeatureSelector interface {
	getFeatures(e Examples) []int
}

type naiveFeatureSelector struct{}

func (n naiveFeatureSelector) getFeatures(e Examples) []int {
	return e.getFeatures()
}

type randomForestFeatureSelector struct {
	featureSampleSize int
}

func (r randomForestFeatureSelector) getFeatures(e Examples) []int {
	features := e.getFeatures()
	perm := rand.Perm(len(features))

	// sampleSize = min(feature sample size, num features)
	sampleSize := r.featureSampleSize
	if sampleSize > len(features) {
		sampleSize = len(features)
	}

	result := make([]int, 0, sampleSize)
	for i := 0; i < sampleSize; i++ {
		result = append(result, features[perm[i]])
	}
	return result
}
