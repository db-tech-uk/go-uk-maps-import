package sqlwriter

import (
	"github.com/db-tech-uk/go-progress/progress"
)

func FinishSQLFiles(sqlFiles []string) error {
	funcName := "sqlwriter.FinishSQLFiles"
	jobName := "Finishing SQL Files"

	magnitude := len(sqlFiles)

	// Finish SQL Files Job
	var job progress.ProgressJob = &FinishSQLFilesJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, sqlFiles, struct{}{})
}
