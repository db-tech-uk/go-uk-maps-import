package database

import (
	"fmt"
	"sort"
	"time"

	"github.com/db-tech-uk/go-logger/logger"
	"github.com/db-tech-uk/go-progress/progress"
	"github.com/db-tech-uk/go-utils/fileutils"
	"github.com/db-tech-uk/go-utils/timeutils"

	"go-uk-maps-import/database/engine"
	"go-uk-maps-import/database/types"
)

type TablesInfo map[string]int

func (t TablesInfo) String() string {
	var r string

	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		if t[k] > 0 {
			r += fmt.Sprintf("\n\t[%v] %v", k, t[k])
		}
	}

	return r
}

type TableCountsResult struct {
	TotalRows   int
	TableCounts TablesInfo
}

func GetTableCounts(s engine.StorageEngine) (TableCountsResult, error) {
	funcName := "database.GetTableCounts"
	jobName := "Counting database rows"

	start := time.Now()
	var took time.Duration
	magnitude := len(types.MapLayers)

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("%v [%v]\n", jobName, magnitude),
	)

	// Table Counts Job
	var j progress.ProgressJob = &TableCountsJob{}

	job, err := j.Setup(jobName, s)
	if err != nil {
		return TableCountsResult{}, fmt.Errorf("%v: %v", funcName, err.Error())
	}
	defer job.End(true)
	res, err := j.Run(job, s)
	if err != nil {
		return TableCountsResult{}, fmt.Errorf("%v: %v", funcName, err.Error())
	}

	took = timeutils.Took(start)
	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("Done %v [%v]\n", jobName, took),
	)

	if r, ok := res.(TableCountsResult); ok {
		return r, nil
	}

	return TableCountsResult{}, fmt.Errorf("incorrect result from job %v", res)
}

func GetGeneratedSQLFiles(folder string) ([]string, error) {
	var sqlFiles []string

	sqlFiles, err := fileutils.Find(folder, ".sql")
	if err != nil {
		return []string{}, err
	}

	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("%v SQL files generated\n", len(sqlFiles)),
	)

	return sqlFiles, nil
}
