package main

import (
	"encoding/json"
	"flag"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"io/ioutil"
	"log"
)

func parseConfig(filename string) (*pb.ForestConfig, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &pb.ForestConfig{}
	err = json.Unmarshal(f, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

var (
	trainFile  = flag.String("train_file", "census-income-test.txt", "")
	testFile   = flag.String("test_file", "census-income-train.txt", "")
	configFile = flag.String("config_file", "forest_config.json", "")
)

func main() {
	// deserialize config
	flag.Parse()
	forestConfig, err := parseConfig(*configFile)
	if err != nil {
		log.Fatal("Failed to load config")
	}

	log.Println(forestConfig.String())
}
