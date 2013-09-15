package mongotrainer

import (
	"flag"
	"time"
)

var (
	// MongoServer to connect to
	MongoServer = flag.String("mongo_server", "localhost", "")
	// MongoDatabase to use
	MongoDatabase = flag.String("mongo_database", "test", "")
	// MongoCollection with active tasks
	MongoCollection = flag.String("mongo_collection", "decisiontrees", "")
	// MongoFs with GridFS files
	MongoFs = flag.String("mongo_fs", "fs", "")
	// MongoPollTime is the interval between polling the task collection for
	// unclaimed tasks
	MongoPollTime = flag.Duration("mongo_poll_time", 30*time.Second, "")
)
