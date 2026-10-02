/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"strings"

	_ "github.com/lib/pq"

	"github.com/spf13/cobra"
	"gov.gsa.fac.cgov-util/internal/logging"
	"gov.gsa.fac.cgov-util/internal/pipes"
	"gov.gsa.fac.cgov-util/internal/structs"
	"gov.gsa.fac.cgov-util/internal/util"

	"gov.gsa.fac.cgov-util/internal/vcap"
)

var (
	db_to_s3_s3path string
	db_to_s3_db     string
)

func get_table_and_schema_names(source_creds vcap.Credentials) map[string]string {
	// Do this table-by-table for RAM reasons.
	db, err := sql.Open("postgres", source_creds.Get("uri").String())
	if err != nil {
		logging.Logger.Println("DBTOS3 could not connect to DB for table-by-table dump")
		logging.Logger.Printf("DBTOS3 %s\n", err)
		os.Exit(logging.DB_SCHEMA_SCAN_FAILURE)
	}

	tables, err := db.Query("SELECT schemaname, tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		logging.Logger.Println("DBTOS3 could not get table names for table-by-table dump")
		logging.Logger.Printf("DBTOS3 %s\n", err)
		os.Exit(logging.DB_SCHEMA_SCAN_FAILURE)
	}

	table_names := make(map[string]string, 0)

	for tables.Next() {
		var table string
		var schema string
		if err := tables.Scan(&schema, &table); err != nil {
			logging.Logger.Println("DBTOS3 could not scan table names in SELECT")
			os.Exit(logging.DB_SCHEMA_SCAN_FAILURE)
		}
		table_names[table] = schema
	}

	return table_names
}

func tables_to_local_bucket(
	source_creds vcap.Credentials,
	up_creds vcap.Credentials,
	s3path *structs.S3Path,
	table_names []string) {
	var BACKUP_ALL = len(table_names) == 0

	logging.Logger.Printf("DBTOS3 backing up from %s to %s%s\n",
		source_creds.Get("name").String(),
		s3path.Bucket,
		s3path.Key,
	)
	table_to_schema := get_table_and_schema_names(source_creds)

	for table, schema := range table_to_schema {
		// Back up tables under two conditions:
		// 1. When it is in a list of names we want backed up, or
		// 2. When there are no names in the list (backup all).
		if slices.Contains(table_names, table) || BACKUP_ALL {
			mc_pipe := pipes.McWrite(
				pipes.PG_Dump_Table(source_creds, schema, table, "--format c"),
				up_creds,
				fmt.Sprintf("%s%s/%s-%s.dump", s3path.Bucket, s3path.Key, schema, table),
			)
			stdout, _ := mc_pipe.String()
			if strings.Contains(stdout, "ERR") {
				logging.Logger.Println("DBTOS3 `dump | mc` pipe failed")
				os.Exit(logging.PIPE_FAILURE)
			}
		}
	}

}

func tables_to_cgov_bucket(
	source_creds vcap.Credentials,
	s3_creds vcap.Credentials,
	s3path *structs.S3Path,
	table_names []string) {
	var BACKUP_ALL = len(table_names) == 0

	table_to_schema := get_table_and_schema_names(source_creds)
	for table, schema := range table_to_schema {
		if slices.Contains(table_names, table) || BACKUP_ALL {
			s3_pipe := pipes.S3Write(
				pipes.PG_Dump_Table(source_creds, schema, table, "--format c"),
				s3_creds,
				fmt.Sprintf("%s%s/%s-%s.dump", s3path.Bucket, s3path.Key, schema, table),
			)
			s3_pipe.Wait()
			if err := s3_pipe.Error(); err != nil {
				logging.Logger.Printf("DBTOS3 `dump | s3` pipe failed for %s.%s: %v", schema, table, err)
				os.Exit(logging.PIPE_FAILURE)
			}
		}
	}
}

// dumpDbToS3Cmd represents the dumpDbToS3 command
var DbToS3Cmd = &cobra.Command{
	Use:   "db_to_s3",
	Args:  cobra.ArbitraryArgs,
	Short: "Dumps a full database to a file in S3",
	Long: `Dumps a full database to a file in S3
Takes 0 or more table names as arguments. If no arguments are
provided, all tables are backed up.
	`,
	Run: func(cmd *cobra.Command, table_names []string) {
		util.UnsetProxy()
		s3path := parseS3Path(db_to_s3_s3path)
		db_creds := getDBCredentials(db_to_s3_db)
		bucket_creds := getBucketCredentials(s3path)

		ch := structs.Choice{
			Local: func() {
				tables_to_local_bucket(db_creds, bucket_creds, s3path, table_names)
			},
			Remote: func() {
				tables_to_cgov_bucket(db_creds, bucket_creds, s3path, table_names)
			}}
		runLocalOrRemote(ch)
	},
}

func init() {
	rootCmd.AddCommand(DbToS3Cmd)
	parseFlags("db_to_s3", DbToS3Cmd)

}
