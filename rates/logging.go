package rates

import (
	"fmt"
	"io"
	"time"

	"github.com/db-tech-uk/go-logger/logger"

	"go-uk-maps-import/filelogger"
)

func LogRecordsProcessed(logFile io.Writer, rateInfo []RateInfo) {
	processed := CalcRecordsProcessed(rateInfo)

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("%v records processed\n", processed),
	)

	filelogger.Log(
		filelogger.LogLine{
			File: logFile,
			Line: fmt.Sprintf("\n%v records processed", processed),
		},
	)
}

func LogRowsGenerated(logFile io.Writer, rateInfo []RateInfo) {
	generated := CalcRowsGenerated(rateInfo)

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("%v rows generated\n", generated),
	)

	filelogger.Log(
		filelogger.LogLine{
			File: logFile,
			Line: fmt.Sprintf("%v rows generated\n", generated),
		},
	)
}

func LogAvgRate(logFile io.Writer, rateInfo []RateInfo) {
	avgRate := CalcAvgRate(rateInfo)

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("Avg rate %v records/s", avgRate),
	)

	filelogger.Log(
		filelogger.LogLine{
			File: logFile,
			Line: fmt.Sprintf("avg rate %v records/s", avgRate),
		},
	)
}

func LogActualRate(logFile io.Writer, rateInfo []RateInfo, duration time.Duration) {
	actualRate := CalcActualRate(rateInfo, duration)

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("Actual rate %v records/s", actualRate),
	)

	filelogger.Log(
		filelogger.LogLine{
			File: logFile,
			Line: fmt.Sprintf("actual rate %v records/s", actualRate),
		},
	)
}

func LogMaxRate(logFile io.Writer, rateInfo []RateInfo) {
	actualRate := CalcMaxRate(rateInfo)

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("Max rate %v records/s", actualRate),
	)

	filelogger.Log(
		filelogger.LogLine{
			File: logFile,
			Line: fmt.Sprintf("max rate %v records/s", actualRate),
		},
	)
}
