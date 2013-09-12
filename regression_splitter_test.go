package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"flag"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"math"
	"math/rand"
	"testing"
)

func constructSmallExamples(numExamples int, numFeatures int) Examples {
	result := make([]*Example, 0, numExamples)
	for i := 0; i < numExamples; i++ {
		example := &Example{
			Features: make([]float64, numFeatures),
		}
		sample := rand.NormFloat64()
		example.Features[rand.Intn(numFeatures)] = sample

		if sample < 0.5 {
			example.Label = 1.0
			example.WeightedLabel = 1.0
		} else {
			example.Label = -1.0
			example.WeightedLabel = -1.0
		}
		result = append(result, example)
	}
	return result
}

// Tests that we split correctly on a trivial example
// label == f[0] > 0.5
func TestBestSplit(t *testing.T) {
	examples := []*Example{
		{
			Features:      []float64{0.0},
			Label:         0.0,
			WeightedLabel: 0.0,
		},
		{
			Features:      []float64{1.0},
			Label:         1.0,
			WeightedLabel: 1.0,
		},
		{
			Features:      []float64{1.0},
			Label:         1.0,
			WeightedLabel: 1.0,
		},
		{
			Features:      []float64{0.0},
			Label:         0.0,
			WeightedLabel: 0.0,
		},
	}
	bestSplit := getBestSplit(examples, 0 /* feature */)
	if bestSplit.feature != 0 {
		t.Fatal(bestSplit)
	}
	if bestSplit.index != 2 {
		t.Fatal(bestSplit)
	}
	if math.Abs(bestSplit.gain-1.0) > 0.001 {
		t.Fatal(bestSplit)
	}
}

func TestRegressionSplitter(t *testing.T) {
	examples := constructSmallExamples(5, 5)
	rs := &RegressionSplitter{
		lossFunction: LogitLoss{
			evaluator: EvaluatorFunc(func(features []float64) float64 { return 0.5 }),
		},
		splittingConstraints: &pb.SplittingConstraints{
			MaximumLevels: proto.Int64(3),
		},
		shrinkageConfig: &pb.ShrinkageConfig{},
	}

	tree := rs.GenerateTree(examples)
	t.Logf("Tree: %+v", tree)
}

func constructBenchmarkExamples(numExamples int, numFeatures int, threshold float64) Examples {
	glog.Info("Num examples: ", numExamples)
	result := make([]*Example, 0, numExamples)
	for i := 0; i < numExamples; i++ {
		example := &Example{
			Features: make([]float64, numFeatures),
		}
		sum := 0.0
		for j := 0; j < numFeatures; j++ {
			sample := rand.NormFloat64()
			sum += sample
			example.Features[int64(j)] = sample
		}
		if sum < threshold {
			example.Label = 1.0
		} else {
			example.Label = -1.0
		}
		result = append(result, example)
	}
	return result
}

func BenchmarkRegressionSplitter(b *testing.B) {
	flag.Parse()

	forestConfig := &pb.ForestConfig{
		NumWeakLearners: proto.Int64(int64(*numTrees)),
		SplittingConstraints: &pb.SplittingConstraints{
			MaximumLevels: proto.Int64(int64(*numLevels)),
		},
		LossFunctionConfig: &pb.LossFunctionConfig{
			LossFunction: pb.LossFunction_LOGIT.Enum(),
		},
	}

	glog.Info(forestConfig.String())

	generator := NewBoostingTreeGenerator(forestConfig)
	examples := constructBenchmarkExamples(b.N, *numFeatures, 0)
	glog.Infof("Starting with %v examples", len(examples))

	b.ResetTimer()
	forest := generator.ConstructBoostingTree(examples)
	res, err := json.MarshalIndent(forest, "", "  ")
	if err != nil {
		glog.Fatalf("Error: %v", err)
	}
	glog.Info(res)
}
