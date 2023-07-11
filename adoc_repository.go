package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ServiceWeaver/weaver"
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
	files, err := ioutil.ReadDir("adocs/")
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
	files, err := ioutil.ReadDir("work/")
	if err != nil {
		return nil, err
	}

	var variants []FileVariant
	for _, file := range files {
		if !file.IsDir() {
			fileNameWithTimestamp := file.Name()
			if strings.HasPrefix(fileNameWithTimestamp, fileName) && strings.HasSuffix(fileNameWithTimestamp, ".adoc") {
				variation := strings.TrimSuffix(strings.TrimPrefix(fileNameWithTimestamp, fileName), ".adoc")

				// Extract the date from the filename suffix
				timestamp := strings.TrimPrefix(variation, "_")
				t, err := time.Parse("20060102_150405", timestamp)
				if err != nil {
					return nil, err
				}

				fileVariant := FileVariant{
					FileName: "work/" + fileNameWithTimestamp,
					Date:     t,
				}
				variants = append(variants, fileVariant)
			}
		}
	}

	return variants, nil
}

func (a *aDocRepository) ReadFile(_ context.Context, fileName string) ([]byte, error) {
	filePath := "adocs/" + fileName + ".adoc"

	variants, err := a.GetVariantionsForFile(context.TODO(), fileName)
	if err != nil {
		return nil, err
	}
	if len(variants) > 0 {
		// Sort the variations by date in descending order
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
	timestamp := time.Now().Format("20060102_150405")
	variantFileName := fmt.Sprintf("%s_%s.adoc", fileName, timestamp)

	filePath := filepath.Join("work", variantFileName)

	err := ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
