/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"golang.org/x/exp/slices"

	"github.com/spf13/cobra"
	"gov.gsa.fac.cgov-util/internal/logging"
	"gov.gsa.fac.cgov-util/internal/pipes"
	"gov.gsa.fac.cgov-util/internal/util"

	vcap "gov.gsa.fac.cgov-util/internal/vcap"
)

var backup_tag string

// func bucket_local(source_creds *vcap.CredentialsRDS, up vcap.UserProvidedCredentials) {
// 	mc_pipe := pipes.Mc(
// 		pipes.PG_Dump(source_creds),
// 		up,
// 		"LOCAL",
// 		"local_db",
// 	)
// 	mc_pipe.Wait()
// 	if err := mc_pipe.Error(); err != nil {
// 		logging.Logger.Println("BACKUPS `dump | mc` pipe failed")
// 		os.Exit(-1)
// 	}
// }

func bucket_local_tables(source_creds *vcap.CredentialsRDS, up vcap.UserProvidedCredentials) {
	table_to_schema := util.Get_table_and_schema_names(source_creds)
	for table, schema := range table_to_schema {
		mc_pipe := pipes.Mc(
			pipes.PG_Dump_Table(source_creds, schema, table, Debug),
			up,
			backup_tag,
			source_creds.DB_Name,
			schema, table, Debug,
		)
		mc_pipe.Wait()
		if err := mc_pipe.Error(); err != nil {
			logging.Logger.Println("BACKUPS `dump | mc` pipe failed")
			os.Exit(-1)
		}
	}
}

func bucket_cgov_tables(source_creds *vcap.CredentialsRDS, up *vcap.CredentialsS3) {
	table_to_schema := util.Get_table_and_schema_names(source_creds)
	for table, schema := range table_to_schema {
		s3_pipe := pipes.S3(
			pipes.PG_Dump_Table(source_creds, schema, table, Debug),
			up,
			backup_tag,
			source_creds.DB_Name,
			schema, table, Debug,
		)
		s3_pipe.Wait()
		if err := s3_pipe.Error(); err != nil {
			logging.Logger.Println("BACKUPS `dump | s3` pipe failed")
			os.Exit(-1)
		}
	}
}

// bucketCmd represents the bucket command
var bucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		source_creds, _ := vcap.GetRDSCreds(SourceDB, "")
		if slices.Contains([]string{"LOCAL", "TESTING"}, os.Getenv("ENV")) {
			up, _ := vcap.GetUserProvidedCredentials("mc")
			bucket_local_tables(source_creds, up)
		} else {
			up, _ := vcap.GetS3Credentials(DestinationBucket)
			bucket_cgov_tables(source_creds, up)
		}

	},
}

func init() {
	rootCmd.AddCommand(bucketCmd)
	bucketCmd.Flags().StringVarP(&SourceDB, "source-db", "", "", "source database (req)")
	bucketCmd.Flags().StringVarP(&DestinationBucket, "destination-bucket", "", "", "destination database (req)")
	bucketCmd.Flags().StringVarP(&backup_tag, "backup-tag", "", "", "SNAPSHOT, HOURLY-03, etc. (req)")
	bucketCmd.Flags().BoolVarP(&Debug, "debug", "d", false, "Log debug statements")
	bucketCmd.MarkFlagRequired("source-db")
	bucketCmd.MarkFlagRequired("destination-bucket")
	bucketCmd.MarkFlagRequired("backup_tag")

}
