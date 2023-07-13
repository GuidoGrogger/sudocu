package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ServiceWeaver/weaver"
)

const (
	workDirName  = "work"
	adocsDirName = "adocs"
)

type ADocRepository interface {
	GetFiles(context.Context) ([]string, error)
	ReadFile(context.Context, string) ([]byte, error)
	SaveVariantForFile(context.Context, string, []byte) error
}

type aDocRepository struct {
	weaver.Implements[ADocRepository]
}

func (a *aDocRepository) GetFiles(_ context.Context) ([]string, error) {
	files, err := ioutil.ReadDir(adocsDirName + "/")
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".adoc") {
			fileName := strings.TrimSuffix(file.Name(), ".adoc")
			fileNames = append(fileNames, fileName)
		}
	}

	return fileNames, nil
}

type FileVariant struct {
	FileName string
	Date     time.Time
}

func (a *aDocRepository) GetVariantionsForFile(_ context.Context, fileName string) ([]FileVariant, error) {
	if err := a.ensureWorkDirExists(); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(workDirName + "/")
	if err != nil {
		return nil, err
	}

	var variants []FileVariant
	for _, file := range files {
		if !file.IsDir() {
			fileNameWithTimestamp := file.Name()
			if strings.HasPrefix(fileNameWithTimestamp, fileName) && strings.HasSuffix(fileNameWithTimestamp, ".adoc") {
				variation := strings.TrimSuffix(strings.TrimPrefix(fileNameWithTimestamp, fileName), ".adoc")

				timestamp := strings.TrimPrefix(variation, "_")
				t, err := time.Parse("20060102_150405", timestamp)
				if err != nil {
					return nil, err
				}

				fileVariant := FileVariant{
					FileName: workDirName + "/" + fileNameWithTimestamp,
					Date:     t,
				}
				variants = append(variants, fileVariant)
			}
		}
	}

	return variants, nil
}

func (a *aDocRepository) ReadFile(_ context.Context, fileName string) ([]byte, error) {
	filePath := adocsDirName + "/" + fileName + ".adoc"

	variants, err := a.GetVariantionsForFile(context.TODO(), fileName)
	if err != nil {
		return nil, err
	}
	if len(variants) > 0 {
		sort.Slice(variants, func(i, j int) bool {
			return variants[i].Date.After(variants[j].Date)
		})

		newestVariant := variants[0]
		filePath = newestVariant.FileName
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (a *aDocRepository) SaveVariantForFile(_ context.Context, fileName string, data []byte) error {
	if err := a.ensureWorkDirExists(); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	variantFileName := fmt.Sprintf("%s_%s.adoc", fileName, timestamp)

	filePath := filepath.Join(workDirName, variantFileName)

	err := ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (a *aDocRepository) ensureWorkDirExists() error {
	if _, err := os.Stat(workDirName); os.IsNotExist(err) {
		err = os.Mkdir(workDirName, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
