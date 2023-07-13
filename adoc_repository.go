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

func (a *aDocRepository) GetFiles(ctx context.Context) ([]string, error) {
	files, err := ioutil.ReadDir(adocsDirName + "/")
	if err != nil {
		return nil, err
	}

	return a.filterAdocFiles(files), nil
}

func (a *aDocRepository) filterAdocFiles(files []os.FileInfo) []string {
	var fileNames []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".adoc") {
			fileName := strings.TrimSuffix(file.Name(), ".adoc")
			fileNames = append(fileNames, fileName)
		}
	}
	return fileNames
}

type FileVariant struct {
	FileName string
	Date     time.Time
}

func (a *aDocRepository) GetVariantionsForFile(ctx context.Context, fileName string) ([]FileVariant, error) {
	if err := a.ensureWorkDirExists(); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(workDirName + "/")
	if err != nil {
		return nil, err
	}

	return a.filterFileVariants(files, fileName), nil
}

func (a *aDocRepository) filterFileVariants(files []os.FileInfo, fileName string) []FileVariant {
	var variants []FileVariant
	for _, file := range files {
		if !file.IsDir() {
			fileNameWithTimestamp := file.Name()
			if strings.HasPrefix(fileNameWithTimestamp, fileName) && strings.HasSuffix(fileNameWithTimestamp, ".adoc") {
				variation := strings.TrimSuffix(strings.TrimPrefix(fileNameWithTimestamp, fileName), ".adoc")

				timestamp := strings.TrimPrefix(variation, "_")
				t, err := time.Parse("20060102_150405", timestamp)
				if err != nil {
					continue // skip the variant if timestamp is not valid
				}

				fileVariant := FileVariant{
					FileName: workDirName + "/" + fileNameWithTimestamp,
					Date:     t,
				}
				variants = append(variants, fileVariant)
			}
		}
	}
	return variants
}

func (a *aDocRepository) ReadFile(ctx context.Context, fileName string) ([]byte, error) {
	filePath := adocsDirName + "/" + fileName + ".adoc"

	variants, err := a.GetVariantionsForFile(ctx, fileName)
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

	return ioutil.ReadFile(filePath)
}

func (a *aDocRepository) SaveVariantForFile(ctx context.Context, fileName string, data []byte) error {
	if err := a.ensureWorkDirExists(); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	variantFileName := fmt.Sprintf("%s_%s.adoc", fileName, timestamp)

	filePath := filepath.Join(workDirName, variantFileName)

	return ioutil.WriteFile(filePath, data, 0644)
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
