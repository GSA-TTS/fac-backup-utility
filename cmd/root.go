/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// For S3 operations
	// db       string
	// s3path   string
	// truncate []string

	// // For db-to-db operations
	// source_db      string
	// destination_db string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "cgov-util",
		Short: "A cgov multitool",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(args)
		},
	}
)

func parseFlags(cmd_name string, cmd *cobra.Command) {
	switch cmd_name {
	case "s3_to_db":
		fmt.Println("RUNNING S3_TO_DB FLAGS")
		cmd.Flags().StringVarP(&s3_to_db_s3path, "s3path", "", "", "source s3 path")
		cmd.Flags().StringVarP(&s3_to_db_db, "db", "", "", "target database name")
		cmd.MarkFlagRequired("db")
		cmd.MarkFlagRequired("s3path")
	case "db_to_s3":
		fmt.Println("RUNNING DB_TO_S3 FLAGS")
		cmd.Flags().StringVarP(&db_to_s3_db, "db", "", "", "source database name")
		cmd.Flags().StringVarP(&db_to_s3_s3path, "s3path", "", "", "destination path")
		cmd.MarkFlagRequired("db")
		cmd.MarkFlagRequired("s3path")
	case "db_to_db":
		fmt.Println("RUNNING DB_TO_DB FLAGS")
		cmd.Flags().StringVarP(&source_db, "src_db", "", "", "source database name")
		cmd.Flags().StringVarP(&dest_db, "dest_db", "", "", "destination database name")
		cmd.Flags().StringVarP(&operation, "operation", "", "", "operation (initial/backup/restore)")
		cmd.MarkFlagRequired("src_db")
		cmd.MarkFlagRequired("dest_db")
		cmd.MarkFlagRequired("operation")
	case "truncate":
		fmt.Println("RUNNING TRUNCATE FLAGS")
		cmd.Flags().StringVarP(&truncate_db, "db", "", "", "target database name")
		cmd.Flags().StringSliceVarP(&truncate_truncate, "truncate", "", []string{}, "tables to truncate before load")
	case "check_tables":
		fmt.Println("RUNNING CHECK_TABLES FLAGS")
		cmd.Flags().StringVarP(&source_database, "db", "", "", "database name")
		cmd.MarkFlagRequired("source_database")
	case "drop":
		fmt.Println("RUNNING DROP FLAGS")
		cmd.Flags().StringVarP(&target_db, "db", "", "", "target database name")
		cmd.Flags().StringSliceVarP(&drop, "drop", "", []string{}, "tables to drop before backup")
	case "row_count":
		fmt.Println("RUNNING ROW_COUNT FLAGS")
		cmd.Flags().StringVarP(&row_count_db, "db", "", "", "database name")
		cmd.MarkFlagRequired("row_count_db")
	default:
		fmt.Printf("NO FLAGS PROCESSED")
	}

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
