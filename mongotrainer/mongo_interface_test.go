package mongotrainer

import (
	"code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestTaskClaiming(t *testing.T) {
	session, err := mgo.Dial(*MongoServer)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("mongo_interface_test")
	c.Remove(bson.M{})
	entry := pb.TrainingRow{
		TrainingStatus: pb.TrainingStatus_UNCLAIMED.Enum(),
		ForestConfig: &pb.ForestConfig{
			NumWeakLearners: proto.Int64(5),
		},
	}

	err = c.Insert(entry)
	if err != nil {
		t.Fatal(err)
	}

	m := MongoTrainer{Collection: c}
	channel := make(chan *trainingTask)
	result := pb.TrainingRow{}
	go func() { m.pollTasks(channel) }()
	task := <-channel
	t.Logf("Claimed task: id: %v, row: %v", task.objectID, task.row)

	assertState := func(status pb.TrainingStatus) {
		err = c.Find(bson.M{}).One(&result)
		if err != nil {
			t.Fatal(err)
		}

		if result.GetTrainingStatus() != status {
			t.Fatalf("Expected status %v, got %v", status, result.GetTrainingStatus())
		}
	}

	assertState(pb.TrainingStatus_UNCLAIMED)
	err = m.claimTask(task)
	if err != nil {
		t.Fatal(err)
	}
	assertState(pb.TrainingStatus_PROCESSING)
	err = m.finalizeTask(task)
	if err != nil {
		t.Fatal(err)
	}
	assertState(pb.TrainingStatus_FINISHED)
	c.Remove(bson.M{})
}
