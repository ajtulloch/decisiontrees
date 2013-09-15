package mongotrainer

import (
	"code.google.com/p/goprotobuf/proto"
	// "github.com/golang/glog"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestMongoInteraction(t *testing.T) {
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
		panic(err)
	}

	result := pb.TrainingRow{}
	err = c.Find(bson.M{}).One(&result)
	if err != nil {
		panic(err)
	}
	t.Log("Got result: ", result.String())
	if !proto.Equal(&result, &entry) {
		t.Fatal(result, entry)
	}
	c.Remove(bson.M{})
}
