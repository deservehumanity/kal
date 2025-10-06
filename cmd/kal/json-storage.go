package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type JSONStorage struct {
	storagePath string
	activities []Activity
}

func NewJSONStorage(storageDir string) (*JSONStorage, error) {
	dataPath := fmt.Sprintf("%s/%s", storageDir, "bin.json")
	if _, err := os.Stat(dataPath); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(dataPath)
		if err != nil {
			return nil, err
		}
		if _, err = file.Write([]byte{'[', ']'}); err != nil {
			return nil, err
		}
		file.Close()
	}

	return &JSONStorage {
		storagePath: dataPath,
		activities: []Activity{},
	}, nil
}

func (s *JSONStorage) Save(activities []Activity) error {
	s.activities = activities

	activitiesJSON, err := json.MarshalIndent(s.activities, "", "\t")
	if err != nil {
		return err
	}
		
	if err = os.WriteFile(s.storagePath, activitiesJSON, fs.ModeTemporary); err != nil {
		return err
	}

	return nil
}

func (s *JSONStorage) Load() ([]Activity, error) {
	bytes, err := os.ReadFile(s.storagePath)
	if (err != nil) {
		return nil, err
	}
	
	if err = json.Unmarshal(bytes, &s.activities); err != nil {
		return nil, err
	}

	return s.activities, nil
}
