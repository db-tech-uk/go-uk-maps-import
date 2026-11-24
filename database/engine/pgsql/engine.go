package pgsql

import (
	"fmt"
	"time"

	"github.com/db-tech-uk/go-logger/logger"
	"github.com/db-tech-uk/go-progress/progress"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"go-uk-maps-import/database/types"
)

const (
	PgSQLTableParams string = ""
)

type PgSQLConfig struct {
	Host    string
	Port    string
	User    string
	Pass    string
	Schema  string
	Timeout int
}

func (c PgSQLConfig) String() string {
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

func (c PgSQLConfig) DSN() string {
	return fmt.Sprintf("user=%v password=%v host=%v port=%v dbname=%v connect_timeout=%v sslmode=disable",
		c.User,
		c.Pass,
		c.Host,
		c.Port,
		c.Schema,
		c.Timeout,
	)
}

type PgSQL struct {
	Config PgSQLConfig
	DB     *sqlx.DB
}

func (e *PgSQL) Connect() error {
	funcName := "pgsql.Connect"
	taskName := "Connecting to PgSQL database"

	logger.Log(
		logger.LVL_INTERNAL,
		fmt.Sprintf("%v\n", taskName),
	)

	// Configure DSN
	dsn := e.Config.DSN()

	// Connect
	db, err := sqlx.Open("postgres", dsn)
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

func (e PgSQL) Cleardown() error {
	funcName := "pgsql.Cleardown"
	jobName := "Cleardown PgSQL Database"

	magnitude := len(types.MapLayers)

	// Cleardown Job
	var job progress.ProgressJob = &ClearDownJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, struct{}{}, e)
}

func (e PgSQL) Prepare() error {
	funcName := "pgsql.Prepare"

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

func (e PgSQL) CreateDatabases() error {
	funcName := "pgsql.CreateDatabases"
	jobName := "Creating PgSQL Databases"

	magnitude := len(types.MapLayers)

	// Create Databases Job
	var job progress.ProgressJob = &CreateDatabasesJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, struct{}{}, e)
}

func (e PgSQL) CreateTables() error {
	funcName := "pgsql.CreateTables"
	jobName := "Creating PgSQL Tables"

	magnitude := len(types.MapLayers)

	// Create Tables Job
	var job progress.ProgressJob = &CreateTablesJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, struct{}{}, e)
}

func (e PgSQL) GetDB(layerType string) *sqlx.DB {
	return e.DB
}

func (e PgSQL) Stop() error {
	e.DB.Close()
	return nil
}

func (e PgSQL) GetTableName(batchInsertsKey string) string {
	return batchInsertsKey
}

func (e PgSQL) GetTableSQL(fullTableName, tableParams string, fields []string) (string, error) {
	tableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", fullTableName)

	for _, f := range fields {
		if _, ok := types.FieldTypes[f]; !ok {
			return "", fmt.Errorf("unknown field type (%v) %v", fullTableName, f)
		}

		tableSQL += fmt.Sprintf("%s %s,", f, types.FieldTypes[f])
	}

	tableSQL += fmt.Sprintf("ogc_geom geometry DEFAULT NULL, UNIQUE (ID, GRIDREF))%v;", tableParams)

	return tableSQL, nil
}
