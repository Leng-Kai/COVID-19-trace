package db

import (
    "context"
    "log"
	"strings"
    _ "github.com/joho/godotenv/autoload"
    "cloud.google.com/go/bigtable"
)

type Trace struct {
    Class   string
    Place   string
    Time    string
}

type UserTraceList struct {
    Uid         string
    TraceList   []Trace
}

var adminClient *bigtable.AdminClient
var client *bigtable.Client
var tables []string

const tableName = "trace-all"

func sliceContains(list []string, target string) bool {
    for _, s := range list {
        if s == target {
            return true
        }
    }
    return false
}

func InitDB(ac *bigtable.AdminClient, c *bigtable.Client, t []string) {
    adminClient = ac
    client = c
    tables = t
}

func HandleTraceFetch(uid string) (UserTraceList, []error) {

    ctx := context.Background()
	userTraceList := UserTraceList{ Uid: uid }
	var traceList []Trace
    var errs []error

    // Create table if it's not yet created

    if !sliceContains(tables, tableName) {
        log.Printf("Creating table %s", tableName)
        if err := adminClient.CreateTable(ctx, tableName); err != nil {
            log.Fatalf("Could not create table %s: %v", tableName, err)
            return userTraceList, append(errs, err)
        }
    }

    // Fetch

    tbl := client.Open(tableName)
    rowKey := uid

    // for i, trace := range traceList {
    //     muts[i] = bigtable.NewMutation()
    //     columnFamilyName := trace.Class
    //     columnName := fmt.Sprintf("%s#%s", trace.Place, trace.Time)
    //     muts[i].Set(columnFamilyName, columnName, bigtable.Now(), []byte("1"))
    //     rowKeys[i] = uid
    // }

	row, err := tbl.ReadRow(ctx, rowKey)
	if err != nil {
		log.Fatalf("Could not read row: %v", err)
		return userTraceList, append(errs, err)
	}

	for cf, itemList := range row {
		for _, item := range itemList {
			column := item.Column
			place_time := strings.Split(column, "#")
			trace := Trace{ Class: cf, Place: place_time[0], Time: place_time[1] }
			traceList = append(traceList, trace)
		}
	}

    // Apply mutations

	userTraceList.TraceList = traceList
    return userTraceList, errs
}