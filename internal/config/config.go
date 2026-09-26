package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Url      string `json:"db_url"`
	Username string `json:"current_user_name"`
}

func getconfigfilepath() (string, error) {
	file_location, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error finding file in Home Directory")
	}
	file_directory := file_location + "/" + configFileName
	return file_directory, nil
}

func Read() (Config, error) {
	var gatorconfigdata Config
	file_directory, err := getconfigfilepath()
	if err != nil {
		return gatorconfigdata, err
	}

	file_data, err := os.ReadFile(file_directory)
	if err != nil {
		return gatorconfigdata, fmt.Errorf("Error reading file from file location")
	}

	if err := json.Unmarshal(file_data, &gatorconfigdata); err != nil {
		return gatorconfigdata, fmt.Errorf("Error unmarshaling JSON data")
	}
	return gatorconfigdata, nil
}

func (conf Config) SetUser(username string) error {
	conf.Username = username

	file_directory, err := getconfigfilepath()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	os.WriteFile(file_directory, jsonData, 0644)
	return nil
}
