// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package fileManager

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path"
	"strings"
	"time"
)

type FileManager interface {
	RunCleanerByAge()
	RemoveTaskFiles(kindName string, taskName string) error
	WriteTaskFile(kindName string, taskName string, roundNumber int, nodeName string, endTime time.Time, data []byte) error
	GetTaskAllFile(kindName string, taskName string) ([]string, error)
	GetAllFile() ([]string, error)
}

type fileManager struct {
	reportDir     string
	logger        *zap.Logger
	cleanInterval time.Duration
}

var _ FileManager = &fileManager{}

func NewManager(logger *zap.Logger, reportDir string, cleanInterval time.Duration) (FileManager, error) {
	if logger == nil || len(reportDir) == 0 {
		return nil, fmt.Errorf("bad request")
	}

	// create directory if not exist
	if _, err := os.Stat(reportDir); err != nil {
		if os.IsNotExist(err) {
			// try to create it
			if e := os.MkdirAll(reportDir, os.ModePerm); e != nil {
				return nil, fmt.Errorf("failed to create diretory %v, error=%v", reportDir, e)
			}
			logger.Sugar().Infof("succeed to create reportDir directory %v", reportDir)
		} else {
			return nil, fmt.Errorf("failed to check directory %v, error=%v", reportDir, err)
		}
	}

	p := &fileManager{
		reportDir:     reportDir,
		logger:        logger,
		cleanInterval: cleanInterval,
	}
	return p, nil
}

func getTaskFileEndTime(filePath string) (endTime time.Time, err error) {
	name := path.Base(filePath)
	if len(name) == 0 {
		return time.Time{}, fmt.Errorf("failed to get file name")
	}
	v := strings.Split(name, "_")
	if len(v) < 3 {
		return time.Time{}, fmt.Errorf("file name %v is not agent format to get file name", name)
	}

	return time.Parse(time.RFC3339, v[len(v)-1])

}

func (s *fileManager) cleanByAgeOnce() {
	filelist, e := os.ReadDir(s.reportDir)
	if e != nil {
		s.logger.Sugar().Errorf("failed to read directory %s, error=%v", s.reportDir, e)
		return
	}

	for _, item := range filelist {
		if item.IsDir() {
			continue
		}

		if endTime, e := getTaskFileEndTime(item.Name()); e != nil {
			s.logger.Sugar().Warnf("ignore unknow file %v, error=%v", item.Name(), e)
			continue
		} else {
			if time.Now().Before(endTime) {
				continue
			}
		}

		if e := os.RemoveAll(path.Join(s.reportDir, item.Name())); e != nil {
			s.logger.Sugar().Errorf("failed to remove file %v who reach age, error=%v", item.Name(), e)
		} else {
			s.logger.Sugar().Infof("remove file %v who reach age ", item.Name())
		}
	}

}

// remove files by deadline
func (s *fileManager) RunCleanerByAge() {
	// clean files at interval
	s.logger.Sugar().Infof("start task file cleaner at interval %v", s.cleanInterval.String())
	go func() {
		for {
			s.cleanByAgeOnce()
			<-time.After(s.cleanInterval)
		}
	}()
}

func GenerateTaskFileName(kindName string, taskName string, roundNumber int, nodeName string, endTime time.Time) string {
	suffix := endTime.Format(time.RFC3339)
	return fmt.Sprintf("%s_%s_round%d_%s_%s", kindName, taskName, roundNumber, nodeName, suffix)
}

func (s *fileManager) GetAllFile() ([]string, error) {
	filelist, e := os.ReadDir(s.reportDir)
	if e != nil {
		return nil, fmt.Errorf("failed to read directory %s, error=%v", s.reportDir, e)
	}

	fileList := []string{}
	for _, item := range filelist {
		if item.IsDir() {
			continue
		}
		fileList = append(fileList, path.Join(s.reportDir, item.Name()))
	}
	return fileList, nil
}

func (s *fileManager) GetTaskAllFile(kindName string, taskName string) ([]string, error) {
	filelist, e := os.ReadDir(s.reportDir)
	if e != nil {
		return nil, fmt.Errorf("failed to read directory %s, error=%v", s.reportDir, e)
	}

	fileList := []string{}
	for _, item := range filelist {
		if item.IsDir() {
			continue
		}
		if strings.HasPrefix(item.Name(), fmt.Sprintf("%s_%s_round", kindName, taskName)) {
			fileList = append(fileList, path.Join(s.reportDir, item.Name()))
		}
	}
	return fileList, nil
}

func (s *fileManager) RemoveTaskFiles(kindName string, taskName string) error {
	v, e := s.GetTaskAllFile(kindName, taskName)
	if e != nil {
		return e
	}

	failList := []string{}
	for _, m := range v {
		if e := os.RemoveAll(m); e != nil {
			failList = append(failList, m)
			s.logger.Sugar().Errorf("failed to remove file %v for kind %v task %v, error=%v", m, kindName, taskName, e)
		} else {
			s.logger.Sugar().Info("remove file %v for kind %v task %v", m, kindName, taskName)
		}
	}

	if len(failList) > 0 {
		return fmt.Errorf("failed to remove files: %v", failList)
	}
	return nil

}

func (s *fileManager) WriteTaskFile(kindName string, taskName string, roundNumber int, nodeName string, endTime time.Time, data []byte) error {

	name := GenerateTaskFileName(kindName, taskName, roundNumber, nodeName, endTime)
	filePath := path.Join(s.reportDir, name)

	v := NewFileWriter(filePath)
	defer v.Close()
	if _, e := v.Write(data); e != nil {
		s.logger.Sugar().Errorf("failed to write data to %v for kind %v task %v round %v", filePath, kindName, taskName, roundNumber)
		return e
	}
	s.logger.Sugar().Infof("succeed to write data to %v for kind %v task %v round %v", filePath, kindName, taskName, roundNumber)

	return nil
}
