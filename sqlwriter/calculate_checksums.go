package sqlwriter

import (
	"io"

	"github.com/db-tech-uk/go-progress/progress"
)

func CalculateChecksums(sqlFiles []string, logFile io.Writer) error {
	funcName := "sqlwriter.CalculateChecksums"
	jobName := "Calculating checksums"

	magnitude := len(sqlFiles)

	// Calculate Checksums Job
	var job progress.ProgressJob = &CalculateChecksumsJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, sqlFiles, logFile)
}
