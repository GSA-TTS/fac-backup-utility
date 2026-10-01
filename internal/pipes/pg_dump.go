package pipes

import (
	"fmt"
	"strings"

	"github.com/bitfield/script"
	"gov.gsa.fac.cgov-util/internal/logging"
	"gov.gsa.fac.cgov-util/internal/util"
	"gov.gsa.fac.cgov-util/internal/vcap"
)

func PG_Dump_Table(creds vcap.Credentials,
	schema string,
	table string,
	format string) *script.Pipe {

	// Backup from replica if it exists
	uri := creds.Get("uri")
	if replicaUri := creds.Get("replica_uri"); replicaUri.Exists() {
		uri = replicaUri
	}

	// Compose the command as a slice
	cmd := []string{
		util.PGDUMP_path,
		"--no-password",
		"--no-privileges",
		"--no-owner",
		format, // need plain for db_to_db
		"--table",
		fmt.Sprintf("%s.%s", schema, table),
		"--dbname",
		fmt.Sprintf(uri.String()),
	}
	// Combine the slice for printing and execution.
	combined := strings.Join(cmd[:], " ")
	logging.Logger.Printf("BACKUPS "+util.PGDUMP_path+" targeting %s.%s\n", schema, table)
	if util.IsDebugLevel("DEBUG") {
		fmt.Printf("command: %s\n", combined)
	}
	return script.Exec(combined)
}
