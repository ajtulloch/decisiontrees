package main

import (
	mt "github.com/ajtulloch/decisiontrees/mongotrainer"
	"labix.org/v2/mgo"
)

func main() {
	session, err := mgo.Dial(*mt.MongoServer)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(*mt.MongoDatabase).C(*mt.MongoCollection)

	trainer := mt.MongoTrainer{Collection: c}
	trainer.Loop()
}
