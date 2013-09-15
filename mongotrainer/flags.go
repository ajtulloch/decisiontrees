package mongotrainer

import (
	"flag"
)

var (
	mongoServer     = flag.String("mongo_server", "localhost", "")
	mongoDatabase   = flag.String("mongo_database", "test", "")
	mongoCollection = flag.String("mongo_collection", "decisiontrees", "")
	mongoFs         = flag.String("mongo_fs", "fs", "")
)
