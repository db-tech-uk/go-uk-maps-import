package autoconfig

import (
	"fmt"
	"time"

	"github.com/db-tech-uk/go-logger/logger"
	"github.com/db-tech-uk/go-utils/timeutils"

	"go-uk-maps-import/importer"
)

const notChecked = "not checked"

type AppConfig struct {
	SystemDetails   SystemDetails
	PlatformDetail  PlatformDetail
	ImporterConfig  importer.Config
	LogsFolder      string
	TimingsLogFile  string
	ChecksumLogFile string
	DryRun          bool
}

type PlatformDetail struct {
	RequestedEngine       string
	MySQLDetails          MySQLDetails
	PgSQLDetails          PgSQLDetails
	SQLiteDetails         SQLiteDetails
	SufficientCPUCores    *bool
	SufficientFileHandles *bool
	DBIsSlow              *bool
	FsIsSlow              *bool
}

func (p PlatformDetail) String() string {
	sufficientCPUCores := notChecked
	if p.SufficientCPUCores != nil {
		sufficientCPUCores = fmt.Sprintf("%v", *p.SufficientCPUCores)
	}

	sufficientFileHandles := notChecked
	if p.SufficientFileHandles != nil {
		sufficientFileHandles = fmt.Sprintf("%v", *p.SufficientFileHandles)
	}

	dbIsSlow := notChecked
	if p.SufficientCPUCores != nil {
		dbIsSlow = fmt.Sprintf("%v", *p.DBIsSlow)
	}

	fsIsSlow := notChecked
	if p.FsIsSlow != nil {
		fsIsSlow = fmt.Sprintf("%v", *p.FsIsSlow)
	}

	return fmt.Sprintf("\t\t"+"RequestedEngine: %v"+"\n"+
		"\t\t"+"MySQLDetails: \n%v"+"\n"+
		"\t\t"+"PgSQLDetails: \n%v"+"\n"+
		"\t\t"+"SQLiteDetails: \n%v"+"\n"+
		"\t\t"+"SufficientCPUCores: %v"+"\n"+
		"\t\t"+"SufficientFileHandles: %v"+"\n"+
		"\t\t"+"DBIsSlow: %v"+"\n"+
		"\t\t"+"FsIsSlow: %v",
		p.RequestedEngine,
		p.MySQLDetails,
		p.PgSQLDetails,
		p.SQLiteDetails,
		sufficientCPUCores,
		sufficientFileHandles,
		dbIsSlow,
		fsIsSlow,
	)
}

func AutoConfigure(appConfig *AppConfig) error {
	funcName := "autoconfig.AutoConfigure"
	taskName := "App Auto Config"

	start := time.Now()
	var took time.Duration

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("%v\n", taskName),
	)

	// Assign current config to local variables
	userImporterConfig := appConfig.ImporterConfig
	systemDetails := appConfig.SystemDetails
	platformDetail := &appConfig.PlatformDetail

	// Attempt to set necessary system parameters
	err := SetupSystem(userImporterConfig, &systemDetails)
	if err != nil {
		return fmt.Errorf("%v: %v", funcName, err.Error())
	}

	// Disk benchmark
	doDiskBench(&systemDetails)

	// Storage benchmark
	doStorageBench(userImporterConfig, platformDetail)

	// Check / ReWrite the Config
	importerConfig := GetImporterConfig(platformDetail, systemDetails, userImporterConfig)

	// Save AutoConfigured Values
	appConfig.ImporterConfig = importerConfig
	appConfig.SystemDetails = systemDetails

	took = timeutils.Took(start)
	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("Done %v [%v]\n", taskName, took),
	)

	return nil
}
