package mysql

import (
	"fmt"
	"time"

	"github.com/db-tech-uk/go-logger/logger"
	"github.com/db-tech-uk/go-progress/progress"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"go-uk-maps-import/database/types"
)

const (
	MySQLTableParams string = " ENGINE=MyISAM DEFAULT CHARSET=utf8"
)

type MySQLConfig struct {
	Host    string
	Port    string
	User    string
	Pass    string
	Schema  string
	Timeout int
}

func (c MySQLConfig) String() string {
	return fmt.Sprintf("\t\t\t"+"Host: %v"+"\n"+
		"\t\t\t"+"Port: %v"+"\n"+
		"\t\t\t"+"User: %v"+"\n"+
		"\t\t\t"+"Pass: %v"+"\n"+
		"\t\t\t"+"Schema: %v"+"\n"+
		"\t\t\t"+"Timeout: %v",
		c.Host,
		c.Port,
		c.User,
		"****",
		c.Schema,
		c.Timeout,
	)
}

func (c MySQLConfig) DSN() string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?timeout=%vs",
		c.User,
		c.Pass,
		c.Host,
		c.Port,
		c.Schema,
		c.Timeout,
	)
}

type MySQL struct {
	Config MySQLConfig
	DB     *sqlx.DB
}

func (e *MySQL) Connect() error {
	funcName := "mysql.Connect"
	taskName := "Connecting to MySQL database"

	logger.Log(
		logger.LVL_INTERNAL,
		fmt.Sprintf("%v\n", taskName),
	)

	// Configure DSN
	dsn := e.Config.DSN()

	// Connect
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("%v %v", funcName, err.Error())
	}

	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Attach
	e.DB = db

	logger.Log(
		logger.LVL_DEBUG,
		"Connected\n",
	)

	return nil
}

func (e MySQL) Cleardown() error {
	funcName := "mysql.Cleardown"
	jobName := "Cleardown MySQL Database"

	magnitude := len(types.MapLayers)

	// Cleardown Job
	var job progress.ProgressJob = &ClearDownJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, struct{}{}, e)
}

func (e MySQL) Prepare() error {
	funcName := "mysql.Prepare"

	err := e.CreateDatabases()
	if err != nil {
		return fmt.Errorf("%v %v", funcName, err.Error())
	}

	err = e.CreateTables()
	if err != nil {
		return fmt.Errorf("%v %v", funcName, err.Error())
	}

	return nil
}

func (e MySQL) CreateDatabases() error {
	funcName := "mysql.CreateDatabases"
	jobName := "Creating MySQL Databases"

	magnitude := len(types.MapLayers)

	// Create Databases Job
	var job progress.ProgressJob = &CreateDatabasesJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, struct{}{}, e)
}

func (e MySQL) CreateTables() error {
	funcName := "mysql.CreateTables"
	jobName := "Creating MySQL Tables"

	magnitude := len(types.MapLayers)

	// Create Tables Job
	var job progress.ProgressJob = &CreateTablesJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, struct{}{}, e)
}

func (e MySQL) GetDB(layerType string) *sqlx.DB {
	return e.DB
}

func (e MySQL) Stop() error {
	e.DB.Close()
	return nil
}

func (e MySQL) GetTableName(batchInsertsKey string) string {
	return batchInsertsKey
}

func (e MySQL) GetTableSQL(fullTableName, tableParams string, fields []string) (string, error) {
	tableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", fullTableName)

	for _, f := range fields {
		if _, ok := types.FieldTypes[f]; !ok {
			return "", fmt.Errorf("unknown field type (%v) %v", fullTableName, f)
		}

		tableSQL += fmt.Sprintf("`%s` %s,", f, types.FieldTypes[f])
	}

	tableSQL += fmt.Sprintf("`ogc_geom` geometry DEFAULT NULL, PRIMARY KEY (`ID`, `GRIDREF`))%v;", tableParams)

	return tableSQL, nil
}
