package database

import (
	"fmt"
	"time"

	"github.com/db-tech-uk/go-logger/logger"
	"github.com/db-tech-uk/go-sqlbench/benchmark"
	"github.com/db-tech-uk/go-utils/timeutils"
	"github.com/jmoiron/sqlx"
)

func Benchmark(db *sqlx.DB, dbDisplayName, tableParams string, numRows int) benchmark.BenchmarkResult {
	err := warmupEngine(db, dbDisplayName)
	if err != nil {
		return benchmark.BenchmarkResult{
			Err: err,
		}
	}

	return runBenchmark(db, dbDisplayName, tableParams, numRows)
}

func warmupEngine(db *sqlx.DB, dbDisplayName string) error {
	funcName := "database.warmupEngine"
	taskName := fmt.Sprintf("Warming up %v", dbDisplayName)

	start := time.Now()
	var took time.Duration

	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("%v\n", taskName),
	)

	bs := benchmark.BenchmarkSuite{
		WarmUp:     benchmark.Warmup,
		PrintStats: true,
	}

	if err := bs.WarmUp(db); err != nil {
		return fmt.Errorf("%v: %v", funcName, err.Error())
	}

	took = timeutils.Took(start)
	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("Done %v [%v]\n", taskName, took),
	)

	return nil
}

func runBenchmark(db *sqlx.DB, dbDisplayName, tableParams string, numRows int) benchmark.BenchmarkResult {
	taskName := fmt.Sprintf("Benchmarking %v", dbDisplayName)

	start := time.Now()
	var took time.Duration

	logger.Log(
		logger.LVL_APP,
		fmt.Sprintf("%v\n", taskName),
	)

	b := benchmark.Benchmark{
		Name: fmt.Sprintf("%vBench", dbDisplayName),
		N:    numRows,
		Bm:   benchmark.BmWKBExec,
	}

	err := benchmark.SetupGeomBench(db, tableParams)
	if err != nil {
		return benchmark.BenchmarkResult{
			Err: err,
		}
	}

	result := b.Run(db)

	benchmark.TeardownGeomBench(db)

	took = timeutils.Took(start)
	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("Done %v [%v]\n", taskName, took),
	)

	return result
}
