package util

import (
	"os"

	"gov.gsa.fac.cgov-util/internal/environments"
	"gov.gsa.fac.cgov-util/internal/logging"
)

var AWS_path = "NO PATH SET"
var PGDUMP_path = "NO PATH SET"
var PGRESTORE_path = "NO PATH SET"
var PSQL_path = "NO PATH SET"

func SetPaths(env string) {
	switch env {
	case environments.LOCAL:
		fallthrough
	case environments.TESTING:
		AWS_path = "aws"
		PGDUMP_path = "pg_dump"
		PGRESTORE_path = "pg_restore"
		PSQL_path = "psql"
	case environments.SANDBOX:
		fallthrough
	case environments.DEVELOPMENT:
		fallthrough
	case environments.PREVIEW:
		fallthrough
	case environments.STAGING:
		fallthrough
	case environments.PRODUCTION:
		AWS_path = "/home/vcap/app/bin/aws"
		PGDUMP_path = "/home/vcap/deps/0/apt/usr/lib/postgresql/15/bin/pg_dump"
		PGRESTORE_path = "/home/vcap/deps/0/apt/usr/lib/postgresql/15/bin/pg_restore"
		PSQL_path = "/home/vcap/deps/0/apt/usr/lib/postgresql/15/bin/psql"
	default:
		logging.Logger.Println("ENV was not set, paths for executables have not been set.")
		os.Exit(-1)
	}
}
