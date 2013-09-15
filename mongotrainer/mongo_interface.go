package mongotrainer

import (
	// "code.google.com/p/goprotobuf/proto"
	"fmt"
	dt "github.com/ajtulloch/decisiontrees"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

// MongoTrainer polls a MongoDB collection for changes
// and spins up training jobs based on these changes
type MongoTrainer struct {
	Collection *mgo.Collection
}

type trainingTask struct {
	objectID bson.ObjectId
	row      *pb.TrainingRow
}

type idRow struct {
	ID bson.ObjectId "_id,omitempty"
}

func (m *MongoTrainer) pollTasks(c chan *trainingTask) {
	for {
		select {
		case <-time.After(*MongoPollTime):
			id := idRow{}
			err := m.Collection.Find(bson.M{
				"trainingStatus": pb.TrainingStatus_UNCLAIMED.Enum(),
			}).Select(bson.M{
				"_id": 1,
			}).One(&id)
			if err != nil {
				glog.Error(err)
				continue
			}

			t := &pb.TrainingRow{}
			err = m.Collection.FindId(id.ID).One(t)
			if err != nil {
				glog.Error(err)
				continue
			}
			c <- &trainingTask{
				objectID: id.ID,
				row:      t,
			}
		}
	}
}

// Loop starts the polling thread, and selects on the channel of
// potential tasks
func (m *MongoTrainer) Loop() {
	taskChannel := make(chan *trainingTask)
	go func() { m.pollTasks(taskChannel) }()
	for {
		select {
		case task := <-taskChannel:
			err := m.runTask(task)
			glog.Info("Starting task %v", task)
			if err != nil {
				glog.Errorf("Got error %v running task %v", err, task)
				continue
			}
			glog.Info("Successfully trained task %v", task)
		}
	}
}

func (m *MongoTrainer) runTraining(task *trainingTask) (*pb.Forest, error) {
	dataSource, err := NewDataSource(task.row.GetDataSourceConfig(), m.Collection.Database.Session)
	if err != nil {
		return nil, err
	}
	trainingData, err := dataSource.GetTrainingData()
	if err != nil {
		return nil, err
	}

	// TODO(tulloch) - more generic factory (e.g. LambdaMART, RF, Boosting)
	generator := dt.NewBoostingTreeGenerator(task.row.GetForestConfig())
	return generator.ConstructBoostingTree(trainingData.GetTrain()), nil
}

func (m *MongoTrainer) claimTask(task *trainingTask) error {
	return m.cas(task.objectID, pb.TrainingStatus_UNCLAIMED, pb.TrainingStatus_PROCESSING)
}

func (m *MongoTrainer) runTask(task *trainingTask) error {
	err := m.claimTask(task)
	if err != nil {
		return err
	}

	forest, err := m.runTraining(task)
	if err != nil {
		return err
	}

	m.finalizeTask(task, forest)
	if err != nil {
		return err
	}
	return nil
}

// cas atomically compares-and-swaps the given objectId between the given training statuses
func (m *MongoTrainer) cas(objectID bson.ObjectId, from, to pb.TrainingStatus) error {
	newRow := &pb.TrainingRow{}
	changeInfo, err := m.Collection.Find(bson.M{
		"_id":            objectID,
		"trainingStatus": from.Enum(),
	}).Apply(mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"trainingStatus": to.Enum(),
			},
		},
		ReturnNew: true,
	}, newRow)

	if err != nil {
		return err
	}
	if changeInfo.Updated != 1 {
		return fmt.Errorf("failed CAS'ing task %v from state %v to state %v", objectID, from, to)
	}
	glog.Infof("Updated objectId %v from state %v to state %v", objectID, from, to)
	return nil
}

func (m *MongoTrainer) finalizeTask(task *trainingTask, forest *pb.Forest) error {
	err := m.cas(task.objectID, pb.TrainingStatus_PROCESSING, pb.TrainingStatus_FINISHED)
	if err != nil {
		return err
	}
	return m.Collection.UpdateId(task.objectID, bson.M{
		"$set": bson.M{
			"forest": forest,
		},
	})
}
