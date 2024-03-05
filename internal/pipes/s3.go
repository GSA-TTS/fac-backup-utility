package pipes

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitfield/script"
	"gov.gsa.fac.cgov-util/internal/logging"
	"gov.gsa.fac.cgov-util/internal/vcap"
)

// https://bitfieldconsulting.com/golang/scripting
func S3(in_pipe *script.Pipe,
	up *vcap.CredentialsS3,
	prefix string,
	source_db string,
	schema string, table string, debug bool) *script.Pipe {
	os.Setenv("AWS_ACCESS_KEY_ID", up.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", up.SecretAccessKey)
	os.Setenv("AWS_DEFAULT_REGION", up.Region)

	// https://serverfault.com/questions/886562/streaming-postgresql-pg-dump-to-s3
	cmd := []string{
		"aws",
		"s3",
		"cp",
		"-",
		fmt.Sprintf("s3://%s/backups/%s-%s.dump",
			up.Bucket,
			prefix,
			source_db),
	}

	// Combine the slice for printing and execution.
	combined := strings.Join(cmd[:], " ")
	if debug {
		fmt.Printf("command: %s\n", combined)

	}
	logging.Logger.Printf("BACKUPS s3 targeting %s\n", prefix)
	return in_pipe.Exec(combined)
}
