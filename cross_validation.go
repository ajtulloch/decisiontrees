package decisiontrees

import (
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"sync"
)

type crossValidationFunc func(trainingSet Examples, testingSet Examples) interface{}

func runCrossValidation(numFolds int, e Examples, f crossValidationFunc) []interface{} {
	folds := e.crossValidationSamples(numFolds)
	crossValidatedResults := make([]interface{}, numFolds)
	w := sync.WaitGroup{}
	for i := range folds {
		w.Add(1)
		go func(pos int) {
			testingSet := folds[pos]
			trainingSet := make([]*pb.Example, 0, len(e)*(numFolds-1)/numFolds)
			for i := range folds {
				if i != pos {
					trainingSet = append(trainingSet, folds[i]...)
				}
			}

			crossValidatedResults[pos] = f(trainingSet, testingSet)
			w.Done()
		}(i)
	}
	w.Wait()
	return crossValidatedResults
}
