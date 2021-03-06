package main

import (
	"encoding/json"
	"io/ioutil"
)

// This is the environmental variable in the OS that should be se to your preferred
// environment and matched to a configuration file. E.g. NOTECLERK_ENVIRONMENT=production
// will match to './config/config.production.json'.
const Environment = "NOTECLERK_ENVIRONMENT"
const DataRoot = "NOTECLERK_DATA"

// This struct is the model for a JSON configuration file that should be located in
// ./config/config.<environment>.json, where '.' indicates the server root, and where
// <environment> can be any lowercase value so long as the NOTECLERK_ENVIRONMENT environmental variable matches.
type Config struct {
	Version        string
	LogPath        string
	ServerProtocol string
	ServerIp       string
	ServerPort     string
	DbIp           string
	DbPort         string
	DbUsername     string
	DbPassword     string
	DbName         string
	DbSslMode      string
}

// Load the configuration JSON and return the Config struct. See the Config struct to view the fields that the JSON
// configuration file is responsible for persisting.
// RETURN: Config, error
func LoadConfiguration(path string) (c *Config, err error) {

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return &Config{}, NoteClerkErrWrap(err, ErrLoadConfigurationFailsReadFile)
	}

	conf := &Config{}
	err = json.Unmarshal(file, conf)
	if err != nil {
		return &Config{}, NoteClerkErrWrap(err, ErrLoadConfigurationFailsJsonMarshal)
	}

	configIsEmpty := conf.DbSslMode == "" || conf.DbIp == "" || conf.DbPort == "" || conf.ServerPort == "" ||
		conf.ServerProtocol == "" || conf.ServerIp == "" || conf.Version == "" || conf.LogPath == "" ||
		conf.DbName == "" || conf.DbPassword == "" || conf.DbUsername == ""

	if configIsEmpty {
		return &Config{}, NoteClerkErrWrap(err, ErrLoadConfigurationAbortsAfterJsonMarshalDueToEmptyConfig)
	}

	return conf, nil
}
