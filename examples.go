package decisiontrees

import (
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"math/rand"
	"sort"
)

// Examples is a slice of Example elements
type Examples []*pb.Example

func (e Examples) subsampleExamples(samplingRate float64) Examples {
	for i := range e {
		j := rand.Intn(i + 1)
		e[i], e[j] = e[j], e[i]
	}

	return e[:int64(float64(len(e))*samplingRate)]
}

func (e Examples) boostrapExamples(samplingRate float64) Examples {
	sampleSize := int(samplingRate * float64(len(e)))
	result := make([]*pb.Example, 0, sampleSize)
	for i := 0; i < sampleSize; i++ {
		result = append(result, e[rand.Intn(len(e))])
	}
	return result
}

func (e Examples) crossValidationSamples(folds int) []Examples {
	crossValidatedSamples := make([]Examples, folds)
	for i := range crossValidatedSamples {
		crossValidatedSamples[i] = make([]*pb.Example, 0, len(e)/folds)
	}

	// Do a Fischer-Yates shuffle of the input array
	for i := range e {
		j := rand.Intn(i + 1)
		e[i], e[j] = e[j], e[i]
	}

	for i, ex := range e {
		fold := i % len(crossValidatedSamples)
		crossValidatedSamples[fold] = append(crossValidatedSamples[fold], ex)
	}
	return crossValidatedSamples
}

func (e Examples) String() string {
	i := make([]interface{}, 0, len(e))
	for _, ex := range e {
		i = append(i, fmt.Sprintf("%+v", ex))
	}
	return fmt.Sprint(i...)
}

type by func(e1, e2 *pb.Example) bool

func (by by) Sort(examples Examples) {
	es := &exampleSorter{
		examples: examples,
		by:       by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(es)
}

type exampleSorter struct {
	examples Examples
	by       by
}

func (e *exampleSorter) Len() int {
	return len(e.examples)
}

func (e *exampleSorter) Swap(i int, j int) {
	e.examples[i], e.examples[j] = e.examples[j], e.examples[i]
}

func (e *exampleSorter) Less(i int, j int) bool {
	return e.by(e.examples[i], e.examples[j])
}

func (e Examples) getFeatures() []int {
	vals := make(map[int]bool)
	for _, example := range e {
		for k, v := range example.Features {
			if v != 0.0 {
				vals[k] = true
			}
		}
	}
	res := make([]int, 0, len(vals))
	for k := range vals {
		res = append(res, k)
	}
	return res
}
