/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitfield/script"
	"github.com/spf13/cobra"
	"gov.gsa.fac.cgov-util/internal/environments"
	"gov.gsa.fac.cgov-util/internal/logging"
	"gov.gsa.fac.cgov-util/internal/pipes"
	"gov.gsa.fac.cgov-util/internal/structs"
	"gov.gsa.fac.cgov-util/internal/util"
	"gov.gsa.fac.cgov-util/internal/vcap"
)

func bucket_to_local_tables(
	db_creds vcap.Credentials,
	bucket_creds vcap.Credentials,
	s3path *structs.S3Path,
) {
	var PROTECTED_DB = "fac-db"
	table_to_schema := get_table_and_schema_names(db_creds)
	//fmt.Sprintf("%s%s/%s-%s.dump", s3path.Bucket, s3path.Key, schema, table)
	check_if_table_exists(db_creds)
	for table, schema := range table_to_schema {
		dump_file_name := fmt.Sprintf("%s-%s.dump", schema, table)

		exit_code := 0
		mc_copy := pipes.McCopy(bucket_creds, fmt.Sprintf("%s%s/%s", s3path.Bucket, s3path.Key, dump_file_name))
		stdout, _ := mc_copy.String()
		if strings.Contains(stdout, "ERR") {
			logging.Logger.Printf("PGCOPY reported an error\n")
			logging.Logger.Println(stdout)
			exit_code = logging.PIPE_FAILURE
		}

		if s3_to_db_db == PROTECTED_DB {
			logging.Logger.Printf("Protected Database '%s' found to be target database. Aborting...", PROTECTED_DB)
			os.Exit(logging.PROTECTED_DATABASE)
		} else {
			//truncate_tables(db_creds, []string{table})
			drop_tables(db_creds, []string{table})

			pg_restore := pipes.PG_Restore(db_creds, schema, table)
			restoreOut, restoreError := pg_restore.String()
			util.ErrorCheck(restoreOut, restoreError)

			os.Remove(fmt.Sprintf("./pg_dump_tables/%s", dump_file_name))
			logging.Logger.Printf("REMOVING FILE: %s", dump_file_name)

			if exit_code != 0 {
				os.Exit(exit_code)
			}
		}
	}

}

// FIXME: need s3read...
func bucket_to_cgov_tables(
	s3_creds vcap.Credentials,
	db_creds vcap.Credentials,
	s3path *structs.S3Path,
) {
	var PROTECTED_DB = "fac-db"
	table_to_schema := get_table_and_schema_names(db_creds)
	//fmt.Sprintf("%s%s/%s-%s.dump", s3path.Bucket, s3path.Key, schema, table)
	for table, schema := range table_to_schema {
		dump_file_name := fmt.Sprintf("%s-%s.dump", schema, table)

		exit_code := 0
		s3_copy := pipes.S3Copy(s3_creds, fmt.Sprintf("%s%s/%s", s3path.Bucket, s3path.Key, dump_file_name))
		stdout, _ := s3_copy.String()
		if strings.Contains(stdout, "ERR") {
			logging.Logger.Printf("S3COPY reported an error\n")
			logging.Logger.Println(stdout)
			exit_code = logging.PIPE_FAILURE
		}

		if s3_to_db_db == PROTECTED_DB {
			logging.Logger.Printf("Protected Database '%s' found to be target database. Aborting...", PROTECTED_DB)
			os.Exit(logging.PROTECTED_DATABASE)
		} else {
			//truncate_tables(db_creds, []string{table})
			drop_tables(db_creds, []string{table})

			pg_restore := pipes.PG_Restore(db_creds, schema, table)
			restoreOut, restoreError := pg_restore.String()
			util.ErrorCheck(restoreOut, restoreError)
			logging.Logger.Printf("RESTORE of table %s complete.", table)

			os.Remove(fmt.Sprintf("./pg_dump_tables/%s", dump_file_name))
			logging.Logger.Printf("REMOVING FILE: %s", dump_file_name)
			if exit_code != 0 {
				os.Exit(exit_code)
			}
		}
	}
}

// S3toDBCmd represents the S3toDB command
var S3toDBCmd = &cobra.Command{
	Use:   "s3_to_db",
	Args:  cobra.ArbitraryArgs,
	Short: "Restore pg_dump file to database.",
	Long: `This command takes database and s3 path input, determining
	if this is being run locally with minio or on cloud.gov, copies the .dump
	files from the dedicated s3 storage to disk, truncates the target table,
	and then performs a pg_restore on the designated table.`,
	Run: func(cmd *cobra.Command, args []string) {
		util.UnsetProxy()
		path_struct := parseS3Path(s3_to_db_s3path)
		// Check that we can get credentials.
		db_creds, err := vcap.VCS.GetCredentials("aws-rds", s3_to_db_db)
		if err != nil {
			logging.Logger.Printf("S3toDB could not get DB credentials for %s", s3_to_db_db)
			os.Exit(logging.COULD_NOT_FIND_CREDENTIALS)
		}

		switch os.Getenv("ENV") {
		case environments.LOCAL:
			fallthrough
		case environments.TESTING:
			bucket_creds, err := vcap.VCS.GetCredentials("user-provided", path_struct.Bucket)
			if err != nil {
				logging.Logger.Printf("S3TODB could not get minio credentials")
				os.Exit(logging.COULD_NOT_FIND_CREDENTIALS)
			}
			bucket_to_local_tables(db_creds, bucket_creds, path_struct)
			os.Remove("pg_dump_tables")
			logging.Logger.Println("Finished Restore and cleaning residual files/folders.")
		case environments.SANDBOX:
			fallthrough
		case environments.DEVELOPMENT:
			fallthrough
		case environments.PREVIEW:
			fallthrough
		case environments.STAGING:
			fallthrough
		case environments.PRODUCTION:
			bucket_creds, err := vcap.VCS.GetCredentials("s3", path_struct.Bucket)
			if err != nil {
				logging.Logger.Printf("S3toDB could not get s3 credentials")
				os.Exit(logging.COULD_NOT_FIND_CREDENTIALS)
			}
			bucket_to_cgov_tables(bucket_creds, db_creds, path_struct)
			os.Remove("pg_dump_tables")
			logging.Logger.Println("Finished Restore and cleaning residual files/folders.")
		}
	},
}

var (
	s3_to_db_s3path string
	s3_to_db_db     string
)

func init() {
	PG_dump_prep()
	rootCmd.AddCommand(S3toDBCmd)
	parseFlags("s3_to_db", S3toDBCmd)
}

func PG_dump_prep() {
	script.Exec("mkdir -p pg_dump_tables")
}
