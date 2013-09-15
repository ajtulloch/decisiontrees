package mongotrainer

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"io/ioutil"
	"labix.org/v2/mgo"
)

// DataSource is an interface for extracting training data from
// an arbitrary location
// TODO(tulloch) - make this an iterator?
type DataSource interface {
	GetTrainingData() (*pb.TrainingData, error)
}

type gridFsDataSource struct {
	session *mgo.Session
	config  *pb.GridFsConfig
}

func (g *gridFsDataSource) GetTrainingData() (t *pb.TrainingData, err error) {
	fs := g.session.DB(g.config.GetDatabase()).GridFS(g.config.GetCollection())
	file, err := fs.Open(g.config.GetFile())
	if err != nil {
		return
	}
	defer func() { err = file.Close() }()

	t = &pb.TrainingData{}
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	err = proto.Unmarshal(buf, t)
	if err != nil {
		return
	}

	glog.Infof("Got %v training examples, %v test examples for config %v",
		len(t.GetTrain()), len(t.GetTest()), g.config)
	return
}

// NewDataSource is a factory function that returns a new DataSource given
// the input DataSourceConfig.
func NewDataSource(c *pb.DataSourceConfig, s *mgo.Session) (DataSource, error) {
	switch c.GetDataSource() {
	case pb.DataSource_GRIDFS:
		return &gridFsDataSource{
			session: s,
			config:  c.GetGridFsConfig(),
		}, nil
	}

	return nil, fmt.Errorf("unknown data source: %v", c)
}
