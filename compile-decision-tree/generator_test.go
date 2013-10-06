package main

import (
	"code.google.com/p/goprotobuf/proto"
	dt "github.com/ajtulloch/decisiontrees"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"io/ioutil"
	"math"
	"math/rand"
	"os/exec"
	"strconv"
	"testing"
)

var (
	cppDemo = `
#include <dlfcn.h>
#include <vector>
#include <iostream>  

double (*evaluationFunc)(const double*);

int main(int argc, char** argv) {
  void* handle = dlopen(argv[1], RTLD_LOCAL | RTLD_LAZY);
  *(void **)(&evaluationFunc) = dlsym(handle, "evaluate");
  std::vector<double> f(1000, strtod(argv[2], NULL));
  const double result = (*evaluationFunc)(f.data());
  std::cout << result;
  return 0;
}
  `
)

func makeAnnotatedTree(level int, numFeatures int) *pb.TreeNode {
	if level == 0 {
		return &pb.TreeNode{
			LeafValue: proto.Float64(rand.Float64()),
		}
	}
	splittingFeature := rand.Int63n(int64(numFeatures))
	splittingValue := rand.Float64()
	t := &pb.TreeNode{
		Feature:    proto.Int64(splittingFeature),
		SplitValue: proto.Float64(splittingValue),
		Left:       makeAnnotatedTree(level-1, numFeatures),
		Right:      makeAnnotatedTree(level-1, numFeatures),
		Annotation: &pb.Annotation{
			LeftFraction: proto.Float64(rand.Float64()),
		},
	}
	return t
}

func generateCompiledEvaluator(t *testing.T) string {
	cppDemoFile, _ := ioutil.TempFile("/tmp", "codegen_tree_evaluator_cpp")
	cppExecutable, _ := ioutil.TempFile("/tmp", "codegen_tree_evaluator_binary")

	cppDemoFile.WriteString(cppDemo)

	commands := [][]string{
		{
			"mv",
			cppDemoFile.Name(),
			cppDemoFile.Name() + ".cpp",
		},
		{
			"clang++",
			"-O0",
			cppDemoFile.Name() + ".cpp",
			"-o",
			cppExecutable.Name(),
		},
	}

	for _, command := range commands {
		cmd := exec.Command(command[0], command[1:]...)
		err := cmd.Run()
		if err != nil {
			t.Fatal(err, command)
		}
	}
	return cppExecutable.Name()
}

func TestGeneratingCode(t *testing.T) {
	numTrees, numLevels, numFeatures := 5, 2, 3
	forest := &pb.Forest{
		Trees: make([]*pb.TreeNode, 0, numTrees),
	}
	for i := 0; i < numTrees; i++ {
		forest.Trees = append(forest.Trees, makeAnnotatedTree(numLevels, numFeatures))
	}

	evaluatorBinary := generateCompiledEvaluator(t)
	sharedLibrary, err := compileTree(forest)
	if err != nil {
		t.Fatal(err)
	}

	// Test a range of feature values and verify that the
	// correct value is computed each time
	for _, featureValue := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
		featureValueString := strconv.FormatFloat(featureValue, 'f', -1, 64)
		cmd := exec.Command(evaluatorBinary, sharedLibrary, featureValueString)
		result, _ := cmd.Output()
		evaluation, err := strconv.ParseFloat(string(result), 64)
		if err != nil {
			t.Fatal(err)
		}
		eval, err := dt.NewRescaledFastForestEvaluator(forest)
		if err != nil {
			t.Fatal(err)
		}
		fv := make([]float64, 1000)
		for i := range fv {
			fv[i] = featureValue
		}

		interpreted := eval.Evaluate(fv)
		if math.Abs(interpreted-evaluation) > 0.001 {
			t.Fatal(interpreted, evaluation)
		}
		t.Log(interpreted, evaluation)
	}
}
