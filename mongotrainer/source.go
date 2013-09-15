package mongotrainer

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"io/ioutil"
	"labix.org/v2/mgo"
)

// Source is an interface for extracting training data from
// an arbitrary location
// TODO(tulloch) - make this an iterator?
type Source interface {
	GetTrainingData() (*pb.TrainingData, error)
}

type gridFsSource struct {
	session *mgo.Session
	config  *pb.GridFsConfig
}

func (g *gridFsSource) GetTrainingData() (t *pb.TrainingData, err error) {
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

// NewSource is a factory function that returns a new Source given
// the input DataSourceConfig.
func NewSource(c *pb.DataSourceConfig, s *mgo.Session) (Source, error) {
	switch c.GetDataSource() {
	case pb.DataSource_GRIDFS:
		return &gridFsSource{
			session: s,
			config:  c.GetGridFsConfig(),
		}, nil
	}

	return nil, fmt.Errorf("unknown data source: %v", c)
}
