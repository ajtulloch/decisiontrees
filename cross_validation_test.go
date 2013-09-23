package decisiontrees

import (
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"math"
	"math/rand"
	"testing"
)

func TestCrossValidation(t *testing.T) {
	numExamples := 1000
	examples := make([]*pb.Example, 0, numExamples)

	for i := 0; i < numExamples; i++ {
		examples = append(examples, &pb.Example{
			Features: []float64{rand.Float64()},
		})
	}

	average := func(trainingSet, testingSet Examples) float64 {
		sum := 0.0
		for _, ex := range testingSet {
			sum += ex.Features[0]
		}
		return sum / float64(len(testingSet))
	}

	stdDev := func(trainingSet, testingSet Examples) float64 {
		sumSquares := 0.0
		for _, ex := range testingSet {
			sumSquares += ex.Features[0] * ex.Features[0]
		}
		return math.Sqrt(
			(float64(len(testingSet)) / float64(len(testingSet)-1)) * ((1.0 / float64(len(testingSet)) * sumSquares) -
				math.Pow(average(trainingSet, testingSet), 2)))
	}

	crossValidatedAverage :=
		runCrossValidation(10, examples, crossValidationFunc(average))
	if math.Abs(crossValidatedAverage-0.5) > 0.02 {
		t.Fatalf("Expected %v, got %v", 0.5, crossValidatedAverage)
	}

	crossValidatedStdDev :=
		runCrossValidation(10, examples, crossValidationFunc(stdDev))
	if math.Abs(crossValidatedStdDev-math.Sqrt(1.0/12.0)) > 0.01 {
		t.Fatalf("Expected %v, got %v", math.Sqrt(1.0/12.0), crossValidatedStdDev)
	}
}
