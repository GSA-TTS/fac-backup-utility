package cmd

import (
	"net/url"
	"os"
	"path"

	"gov.gsa.fac.cgov-util/internal/environments"
	"gov.gsa.fac.cgov-util/internal/logging"
	"gov.gsa.fac.cgov-util/internal/structs"
	"gov.gsa.fac.cgov-util/internal/vcap"
)

func parseS3Path(s3_path string) *structs.S3Path {
	u, err := url.Parse(s3_path)
	if err != nil {
		logging.Logger.Printf("parseS3Path could not parse s3 path: %s", s3_path)
		os.Exit(logging.S3_PATH_PARSE_ERROR)
	}
	if u.Scheme != "s3" {
		logging.Logger.Printf("parseS3Path does not look like an S3 path (e.g. `s3://`): %s", s3_path)
		os.Exit(logging.S3_PATH_PARSE_ERROR)
	}
	logging.Logger.Println("Host: ", u.Host)
	logging.Logger.Println("Path: ", u.Path)

	return &structs.S3Path{
		//Bucket: filepath.Clean(u.Host),
		//Key:    filepath.Clean(u.Path),
		Bucket: path.Clean(u.Host),
		Key:    path.Clean(u.Path),
	}
}

func runLocalOrRemote(funs structs.Choice) {
	switch os.Getenv("ENV") {
	case environments.LOCAL:
		fallthrough
	case environments.TESTING:
		funs.Local()
	case environments.SANDBOX:
		fallthrough
	case environments.DEVELOPMENT:
		fallthrough
	case environments.PREVIEW:
		fallthrough
	case environments.STAGING:
		fallthrough
	case environments.PRODUCTION:
		funs.Remote()
	default:
		logging.Logger.Printf("LOCALORREMOTE impossible condition")
		os.Exit(-1)
	}
}

func getDBCredentials(db_name string) vcap.Credentials {
	// Check that we can get credentials.
	db_creds, err := vcap.VCS.GetCredentials("aws-rds", db_name)
	if err != nil {
		logging.Logger.Printf("GETDBCREDENTIALS could not get DB credentials for %s", db_name)
		os.Exit(logging.COULD_NOT_FIND_CREDENTIALS)
	}
	return db_creds
}

func getBucketCredentials(s3path *structs.S3Path) vcap.Credentials {
	switch os.Getenv("ENV") {
	case environments.LOCAL:
		fallthrough
	case environments.TESTING:
		bucket_creds, err := vcap.VCS.GetCredentials("user-provided", s3path.Bucket)
		if err != nil {
			logging.Logger.Printf("GetCredentials could not get minio credentials: %s", s3path)
			os.Exit(logging.COULD_NOT_FIND_CREDENTIALS)
		}
		return bucket_creds
	case environments.SANDBOX:
		fallthrough
	case environments.DEVELOPMENT:
		fallthrough
	case environments.PREVIEW:
		fallthrough
	case environments.STAGING:
		fallthrough
	case environments.PRODUCTION:
		bucket_creds, err := vcap.VCS.GetCredentials("s3", s3path.Bucket)
		if err != nil {
			logging.Logger.Printf("DBTOS3 could not get s3 credentials")
			os.Exit(logging.COULD_NOT_FIND_CREDENTIALS)
		}
		return bucket_creds
	default:
		logging.Logger.Printf("DBTOS3 could not get env for bucket credentials")
		os.Exit(logging.COULD_NOT_FIND_CREDENTIALS)
	}

	// We should never get here.
	logging.Logger.Printf("DBTOS3 impossible condition")
	os.Exit(-1)
	return vcap.Credentials{}
}
