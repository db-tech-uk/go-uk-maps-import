package sqlite

import (
	"fmt"

	"github.com/db-tech-uk/go-progress/progress"
	"github.com/db-tech-uk/go-utils/fileutils"

	"go-uk-maps-import/database/types"
)

// ref: https://groups.google.com/g/spatialite-users/c/U2pxp3bwVnY

var targetFolder string = "sql"

func (e SQLite) ExportToMySQLFiles() error {
	funcName := "sqlite.ExportToMySQLFiles"
	jobName := "Exporting SQLite databases to MySQL format files"

	magnitude := len(types.MapLayers)

	// Pre flight
	err := fileutils.MkDir(targetFolder)
	if err != nil {
		return fmt.Errorf("%v: %v", funcName, err.Error())
	}

	err = fileutils.EmptyFolder(targetFolder)
	if err != nil {
		return fmt.Errorf("%v: %v", funcName, err.Error())
	}

	// ExportToMySQLFiles Job
	var job progress.ProgressJob = &ExportToMySQLFilesJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, struct{}{}, e)
}

func (e SQLite) ExportToSQLiteFiles() error {
	funcName := "mysql.ExportToSQLiteFiles"
	jobName := "Exporting SQLite databases to SQLite format files"

	magnitude := len(types.MapLayers)

	// Pre flight
	err := fileutils.MkDir(targetFolder)
	if err != nil {
		return fmt.Errorf("%v: %v", funcName, err.Error())
	}

	err = fileutils.EmptyFolder(targetFolder)
	if err != nil {
		return fmt.Errorf("%v: %v", funcName, err.Error())
	}

	// ExportToMySQLFiles Job
	var job progress.ProgressJob = &ExportToSQLiteFilesJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, struct{}{}, e)
}
