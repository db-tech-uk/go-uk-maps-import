package database

import (
	"fmt"
	"sort"
	"strings"

	"github.com/db-tech-uk/go-logger/logger"
	"github.com/db-tech-uk/go-utils/sliceutils"
	_ "github.com/go-sql-driver/mysql"

	"go-uk-maps-import/database/engine"
	"go-uk-maps-import/rates"
)

type Range struct {
	Min int
	Max int
}

func RowCountsCheck(dataFolder string, ratesInfo []rates.RateInfo, dbCounts map[string]int, getTableNameFunc func(string) string) []string {
	var mismatches = []string{}
	ranges := getExpectedRange(ratesInfo)

	for _, rateInfo := range ratesInfo {
		shapeFileShortName := strings.ReplaceAll(rateInfo.ShapeFile, dataFolder, "")
		var dbName = GetDBNameFromFilename(rateInfo.ShapeFile)

		for sq := range rateInfo.Rows {
			fullTableName := getTableNameFunc(fmt.Sprintf("%v.%v", dbName, sq))
			targetRange := ranges[fullTableName]
			actual := dbCounts[fullTableName]

			if actual < targetRange.Min || actual > targetRange.Max {
				message := fmt.Sprintf(
					"%v: %v expected [Min:%v, Max:%v], actual %v",
					shapeFileShortName,
					fullTableName,
					targetRange.Min,
					targetRange.Max,
					actual,
				)

				mismatches = append(mismatches, message)
			}
		}
	}

	sort.Strings(mismatches)

	return mismatches
}

func LogRowCountChecks(se engine.StorageEngine, datafolder string, rateInfo []rates.RateInfo) {
	var mismatches []string
	msg := "All rowcounts are as expected"

	tableCountsResult, _ := GetTableCounts(se)

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("%v rows in database\n", tableCountsResult.TotalRows),
	)

	logger.Log(
		logger.LVL_INTERNAL,
		fmt.Sprintf("TableCounts: %v\n", tableCountsResult.TableCounts),
	)

	logger.Log(
		logger.LVL_DEBUG,
		"Running rowcount check\n",
	)

	mismatches = RowCountsCheck(datafolder, rateInfo, tableCountsResult.TableCounts, se.GetTableName)
	if len(mismatches) > 0 {
		msg = fmt.Sprintf("Row Counts %v\n", sliceutils.TabList(mismatches))
	}

	logger.Log(
		logger.LVL_DEBUG,
		msg,
	)
}

func getExpectedRange(ratesInfo []rates.RateInfo) map[string]Range {
	var allrows = map[string][]int{}
	var ranges = map[string]Range{}

	for _, rateInfo := range ratesInfo {
		var dbName = GetDBNameFromFilename(rateInfo.ShapeFile)

		for sq, rows := range rateInfo.Rows {
			fullTableName := fmt.Sprintf("%v.%v", dbName, sq)

			_, ok := allrows[fullTableName]
			if !ok {
				allrows[fullTableName] = []int{rows}
			} else {
				allrows[fullTableName] = append(allrows[fullTableName], rows)
			}
		}
	}

	for fullTableName, rows := range allrows {
		ranges[fullTableName] = Range{
			Min: sliceutils.MaxInt(rows),
			Max: sliceutils.SumInt(rows),
		}
	}

	return ranges
}
