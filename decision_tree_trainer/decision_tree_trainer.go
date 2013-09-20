package main

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"flag"
	dt "github.com/ajtulloch/decisiontrees"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
)

var (
	configPath    = flag.String("config", "dt.json", "")
	trainDataPath = flag.String("train_data", "train_data.json", "")
)

func parseToProto(file string, protobuf proto.Message) error {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(f, protobuf)
}

func main() {
	flag.Parse()
	trainData := &pb.TrainingData{}
	if err := parseToProto(*trainDataPath, trainData); err != nil {
		glog.Fatal(err)
	}

	glog.Infof(
		"Loaded %v training examples, %v test examples",
		len(trainData.GetTrain()),
		len(trainData.GetTest()))

	config := &pb.ForestConfig{}
	if err := parseToProto(*configPath, config); err != nil {
		glog.Fatal(err)
	}
	glog.Infof("Loaded forest config %+v", config)

	generator := dt.NewBoostingTreeGenerator(config)
	forest := generator.ConstructBoostingTree(trainData.GetTrain())
	learningCurve := dt.LearningCurve(forest, trainData.GetTest())

	glog.Infof("Learning curve: %+v", learningCurve)

	serializedForest, err := json.MarshalIndent(forest, "", "  ")
	if err != nil {
		glog.Fatal(err)
	}

	os.Stdout.Write(serializedForest)
}
