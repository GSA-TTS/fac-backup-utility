/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gov.gsa.fac.cgov-util/internal/logging"
	"gov.gsa.fac.cgov-util/internal/vcap"
)

var (
	row_count_db string
)

type RowCount struct {
	Table string `json:"Table"`
	Rows  string `json:"Rows"`
}

type connection struct {
	RowsCountConnection []*RowCount `json:"TABLEROWCOUNT"`
}

func check_rows_in_db(source_creds vcap.Credentials) {
	db, err := sql.Open("postgres", source_creds.Get("uri").String())
	if err != nil {
		logging.Logger.Println("TABLECHECK could not connect to DB for checking table existance")
		logging.Logger.Printf("DBTOS3 %s\n", err)
		os.Exit(logging.DB_SCHEMA_SCAN_FAILURE)
	}
	file, err := f.ReadFile("assets/db_tables.txt")
	//print(string(file))
	if err != nil {
		logging.Error.Println(err)
		os.Exit(logging.ROW_COUNT_ERROR)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(file)))
	var row_count_for_tables []string
	for scanner.Scan() {
		query := fmt.Sprintf("SELECT count(*) FROM %s;", scanner.Text())
		rows, err := db.Query(query)
		if err != nil {
			logging.Error.Println(err)
			os.Exit(logging.ROW_COUNT_ERROR)
		}
		defer rows.Close()
		var count int
		for rows.Next() {
			if err := rows.Scan(&count); err != nil {
				logging.Error.Println(err)
				os.Exit(logging.ROW_COUNT_ERROR)
			}
		}
		// Output to stdout on each line for debugging purposes
		// logging.Logger.Printf(fmt.Sprintf("Table: %s | Row Count: %d\n", scanner.Text(), count))
		r := strconv.Itoa(count)
		// Store in row_count_for_tables []string
		row_count_for_tables = append(row_count_for_tables, scanner.Text()+" "+r)
	}

	logging.Logger.Printf("Row count for tables in manifest...")

	joined_tables := strings.Join(row_count_for_tables[:], "\n")
	//logging.Logger.Printf("TABLEROWCOUNT " + joined_tables)

	var rows []*RowCount
	for _, joined_tables := range strings.Split(joined_tables, "\n") {
		if joined_tables != "" {
			s := strings.Split(joined_tables, " ")
			rows = append(rows, &RowCount{Table: s[0], Rows: s[1]})
		}
	}

	// Raw json object of {"Table":"table_name","Rows":"#"}
	raw, _ := json.Marshal(connection{RowsCountConnection: rows})
	logging.Logger.Printf("%s", raw)
	// PrettyPrint json object of {"Table":"table_name","Rows":"#"}
	//pretty, _ := json.MarshalIndent(connection{RowsCountConnection: rows}, "", "    ")
	//logging.Logger.Printf("%s\n", pretty)

	if err := scanner.Err(); err != nil {
		logging.Error.Println(err)
	}
}

// rowCountCmd represents the rowCount command
var rowCountCmd = &cobra.Command{
	Use:   "row_count",
	Short: "Check the rows in a given table",
	Long:  `Check the rows in a given table`,
	Run: func(cmd *cobra.Command, args []string) {
		db_creds := getDBCredentials(row_count_db)
		check_rows_in_db(db_creds)
	},
}

func init() {
	rootCmd.AddCommand(rowCountCmd)
	parseFlags("row_count", rowCountCmd)
}
