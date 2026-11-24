package engine

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/db-tech-uk/go-logger/logger"
	"github.com/db-tech-uk/go-progress/progress"
	"github.com/db-tech-uk/go-utils/osutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"go-uk-maps-import/database/engine/mysql"
	"go-uk-maps-import/database/engine/pgsql"
	"go-uk-maps-import/database/engine/sqlite"
)

const nilPlaceholder = "<nil>"

var (
	EngineMySQL    = "mysql"
	EngineSQLite   = "sqlite"
	EnginePostgres = "pgsql"
)

type StorageEngine interface {
	Connect() error
	Cleardown() error
	Prepare() error
	Stop() error
	GetDB(layerType string) *sqlx.DB
	GetTableName(batchInsertsKey string) string
	GetTableSQL(fullTableName, tableParams string, fields []string) (string, error)
}

type DBConfig struct {
	Host    *string
	Port    *string
	User    *string
	Pass    *string
	Schema  *string
	Timeout *int
}

func (c DBConfig) String() string {
	host := nilPlaceholder
	if c.Host != nil {
		host = *c.Host
	}

	port := nilPlaceholder
	if c.Port != nil {
		port = *c.Port
	}

	user := nilPlaceholder
	if c.User != nil {
		user = *c.User
	}

	schema := nilPlaceholder
	if c.Schema != nil {
		schema = *c.Schema
	}

	var timeout int
	if c.Timeout != nil {
		timeout = *c.Timeout
	}

	return fmt.Sprintf("\t\t\t"+"Host: %v"+"\n"+
		"\t\t\t"+"Port: %v"+"\n"+
		"\t\t\t"+"User: %v"+"\n"+
		"\t\t\t"+"Pass: %v"+"\n"+
		"\t\t\t"+"Schema: %v"+"\n"+
		"\t\t\t"+"Timeout: %v",
		host,
		port,
		user,
		"****",
		schema,
		timeout,
	)
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?timeout=%vs",
		c.User,
		c.Pass,
		c.Host,
		c.Port,
		c.Schema,
		c.Timeout,
	)
}

type SEConfig struct {
	Engine        *string
	DBConfig      DBConfig
	ClearDown     bool
	CountsOnly    bool
	StorageEngine StorageEngine
}

func (c SEConfig) String() string {
	engine := nilPlaceholder
	if c.Engine != nil {
		engine = fmt.Sprintf("%v", *c.Engine)
	}

	return fmt.Sprintf("\t\t"+"Engine: %v"+"\n"+
		"\t\t"+"DBConfig:"+"\n"+"%s"+"\n"+
		"\t\t"+"ClearDown: %v",
		engine,
		c.DBConfig,
		c.ClearDown,
	)
}

func Startup(dryrun bool, config *SEConfig) error {
	funcName := "engine.Startup"

	if !dryrun {
		e, err := start(config)
		if err != nil {
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}
		config.StorageEngine = e
	}

	return nil
}

func Shutdown(dryrun bool, config SEConfig) error {
	funcName := "engine.Shutdown"

	if !dryrun {
		err := config.StorageEngine.Stop()
		if err != nil {
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}
	}

	return nil
}

//nolint:ireturn,nolintlint
func start(config *SEConfig) (StorageEngine, error) {
	funcName := "engine.start"
	var e StorageEngine

	switch *config.Engine {
	case EngineMySQL:
		e = &mysql.MySQL{
			Config: mysql.MySQLConfig{
				Host:    *config.DBConfig.Host,
				Port:    *config.DBConfig.Port,
				User:    *config.DBConfig.User,
				Pass:    *config.DBConfig.Pass,
				Schema:  *config.DBConfig.Schema,
				Timeout: *config.DBConfig.Timeout,
			},
		}

	case EnginePostgres:
		e = &pgsql.PgSQL{
			Config: pgsql.PgSQLConfig{
				Host:    *config.DBConfig.Host,
				Port:    *config.DBConfig.Port,
				User:    *config.DBConfig.User,
				Pass:    *config.DBConfig.Pass,
				Schema:  *config.DBConfig.Schema,
				Timeout: *config.DBConfig.Timeout,
			},
		}

	case EngineSQLite:
		e = &sqlite.SQLite{
			Config: sqlite.SQLiteConfig{},
		}

	default:
		return nil, fmt.Errorf("database engine type not set [%v]", config.Engine)
	}

	// Connect to the database
	err := e.Connect()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", funcName, err.Error())
	}

	// Cleardown the database
	if config.ClearDown {
		err := e.Cleardown()
		if err != nil {
			return nil, fmt.Errorf("%v: %v", funcName, err.Error())
		}
	}

	if !config.CountsOnly {
		// Prep the database
		err := e.Prepare()
		if err != nil {
			return nil, fmt.Errorf("%v: %v", funcName, err.Error())
		}
	}

	return e, nil
}

func DoInserts(dbConfig DBConfig, sqlFiles []string) error {
	funcName := "engine.DoInserts"
	jobName := "Inserting data to the db"

	magnitude := len(sqlFiles)

	// Do Inserts Job
	var job progress.ProgressJob = &DoInsertsJob{}

	return progress.RunJob(jobName, funcName, job, magnitude, sqlFiles, dbConfig)
}

func RunSQLFileDirect(config SEConfig, sqlFile string) (string, error) {
	funcName := "engine.RunSQLFileDirect"

	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("Running SQL file %v \n", sqlFile),
	)

	layerType := GetLayerType(sqlFile)

	db := config.StorageEngine.GetDB(layerType)

	c, err := os.ReadFile(sqlFile)
	if err != nil {
		return "", fmt.Errorf("%v: %v", funcName, err.Error())
	}

	sql := string(c)
	_, err = db.Exec(sql)
	if err != nil {
		return "", fmt.Errorf("%v: %v", funcName, err.Error())
	}

	return "", nil
}

func GetLayerType(sqlFileName string) string {
	s := strings.Split(sqlFileName, "/")
	return s[0]
}

func RunSQLFile(config DBConfig, sqlFile string) (string, string, error) {
	command := "mysql"

	args := []string{
		"--connect-timeout",
		"10",
		fmt.Sprintf("-h%v", config.Host),
		fmt.Sprintf("-P%v", config.Port),
		fmt.Sprintf("-u%v", config.User),
		fmt.Sprintf("-p%v", config.Pass),
		fmt.Sprintf("-D%v", config.Schema),
		"-e",
		fmt.Sprintf("source %v", sqlFile),
	}

	// don't write the password to the log
	logArgs := slices.Clone(args)
	if i := slices.Index(logArgs, fmt.Sprintf("-p%v", config.Pass)); i >= 0 {
		logArgs[i] = "-p****"
	}

	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("Running command: %s %s", command, strings.Join(logArgs, " ")),
	)

	return osutils.RunCommandSilent(command, args...)
}

func RunSQLFileWithShell(shell string, config DBConfig, sqlFile string) (string, string, error) {
	const tpl = "mysql --connect-timeout 10 -h%v -P %v -u %v -p%v %v < %v"

	command := fmt.Sprintf(tpl, config.Host, config.Port, config.User, config.Pass, config.Schema, sqlFile)

	// don't write the password to the log
	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("Running command: %s", fmt.Sprintf(tpl, config.Host, config.Port, config.User, "****", config.Schema, sqlFile)),
	)

	return osutils.RunShellCommand(shell, command)
}
