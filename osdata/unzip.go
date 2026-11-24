package osdata

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/db-tech-uk/csync/waitgroup"
	"github.com/db-tech-uk/go-logger/logger"
	"github.com/db-tech-uk/go-progress/progress"
	"github.com/db-tech-uk/go-utils/fileutils"
	"github.com/db-tech-uk/go-utils/stringutils"
	"github.com/db-tech-uk/uiprogress"
)

func unzipTile(tile VectorMapDistrictTile) error {
	funcName := "osdata.unzipTile"

	destFolder := fmt.Sprintf("%v/%v", shpDir, tile.Area)

	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("unzipping tile [%v]\n", tile.Area),
	)

	if fileutils.FolderExists(destFolder) {
		err := os.RemoveAll(destFolder)
		if err != nil {
			return fmt.Errorf("%v: unable to remove target folder %v [%v]", funcName, destFolder, err.Error())
		}
	}

	folderWithinZip := fmt.Sprintf("OS VectorMap District (ESRI Shape File) %v/data/", tile.Area)
	zipFile := fmt.Sprintf("%v/%v", zipDir, tile.FileName)
	var archive *zip.ReadCloser

	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("%v: %v", funcName, err.Error())
	}
	defer archive.Close()

	magnitude := getFileCount(archive, folderWithinZip)
	fileName := tile.FileName
	wg := waitgroup.New()

	logger.Log(
		logger.LVL_DEBUG,
		fmt.Sprintf("Unzipping %v [%v files]\n", tile.FileName, magnitude),
	)

	task := &progress.Task{
		ID:        fileName,
		Magnitude: float64(magnitude),
	}

	var tasks = []*progress.Task{
		task,
	}

	job := progress.NewJob(fileName, len(tasks))
	job.AddTasks(tasks)

	if progress.ShouldShowBar() {
		uiprogress.Start()

		job.Bar = uiprogress.AddBar(magnitude).PrependCompleted()
		job.Bar.AppendFunc(func(b *uiprogress.Bar) string {
			status, _ := job.GetStatus()
			taskName := fmt.Sprintf("%v %v files", fileName, magnitude)
			return fmt.Sprintf("%s [%s]", stringutils.SpacePadRight(taskName, 22), status)
		})

		defer job.End(false)

		err = job.Start()
		if err != nil {
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}

		task.Start()
		wg.Add(1)

		c := make(chan error)
		go func(j *progress.Job, t *progress.Task, zf string) {
			err := unzip(destFolder, zf, folderWithinZip, j)
			if err != nil {
				c <- err
			}
			t.End()
			wg.Done()

			c <- nil
		}(job, task, zipFile)
		err = <-c
		if err != nil {
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}

		wg.Wait()

		uiprogress.Stop()
	} else {
		err := unzip(destFolder, zipFile, folderWithinZip, job)
		if err != nil {
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}
	}

	return nil
}

func unzip(dst string, zipfile string, folderWithinZip string, j *progress.Job) error {
	funcName := "osdata.unzip"

	var archive *zip.ReadCloser

	archive, err := zip.OpenReader(zipfile)
	if err != nil {
		return fmt.Errorf("%v: %v", funcName, err.Error())
	}
	defer archive.Close()

	for _, f := range archive.File {
		if !strings.HasPrefix(f.Name, folderWithinZip) {
			continue
		}

		if f.FileInfo().IsDir() {
			continue
		}

		filePath := fmt.Sprintf("%v/%v", dst, f.Name)
		destPath := strings.ReplaceAll(filePath, folderWithinZip, "")

		logger.Log(
			logger.LVL_INTERNAL,
			fmt.Sprintf("unzipping file [%v] to [%v]", f.Name, destPath),
		)

		destFolder := fmt.Sprintf("%v/", filepath.Clean(dst))
		if !strings.HasPrefix(strings.TrimPrefix(filePath, "./"), destFolder) {
			return fmt.Errorf("invalid file path %v [%v]", filePath, destFolder)
		}

		err := fileutils.MkDir(destFolder)
		if err != nil {
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}

		dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}
		defer dstFile.Close()

		fileInArchive, err := f.Open()
		if err != nil {
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}

		_, err = io.CopyN(dstFile, fileInArchive, f.FileInfo().Size())
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			return fmt.Errorf("%v: %v", funcName, err.Error())
		}
		defer fileInArchive.Close()

		if progress.ShouldShowBar() {
			j.Bar.Incr()
		}
	}

	return nil
}

func getFileCount(archive *zip.ReadCloser, folderWithinZip string) int {
	var count int

	for _, f := range archive.File {
		if !strings.HasPrefix(f.Name, folderWithinZip) {
			continue
		}

		if f.FileInfo().IsDir() {
			continue
		}

		count++
	}

	return count
}
