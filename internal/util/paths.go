package util

import (
	"os"

	"gov.gsa.fac.cgov-util/internal/logging"
)

var AWS_path = "NO PATH SET"
var PGDUMP_path = "NO PATH SET"
var PGRESTORE_path = "NO PATH SET"
var PSQL_path = "NO PATH SET"

func SetPaths(env string) {
	switch env {
	case "LOCAL":
		fallthrough
	case "TESTING":
		AWS_path = "aws"
		PGDUMP_path = "pg_dump"
		PGRESTORE_path = "pg_restore"
		PSQL_path = "psql"
	case "DEVELOPMENT":
		fallthrough
	case "STAGING":
		fallthrough
	case "PREVIEW":
		fallthrough
	case "PRODUCTION":
		AWS_path = "/home/vcap/app/bin/aws"
		PGDUMP_path = "/home/vcap/deps/0/apt/usr/lib/postgresql/15/bin/pg_dump"
		PGRESTORE_path = "/home/vcap/deps/0/apt/usr/lib/postgresql/15/bin/pg_restore"
		PSQL_path = "/home/vcap/deps/0/apt/usr/lib/postgresql/15/bin/psql"
	default:
		logging.Logger.Println("ENV was not set, paths for executables have not been set.")
		os.Exit(-1)
	}
}
