package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	Hubs map[string]string `json:"hubs"`
	//User string `json:"user"`
	//Email string `json:"email"`
}

func readConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	cfg := new(Config)
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}
	// TODO: verify data.
	return cfg, nil
}

func writeConfig(file string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	if err := ioutil.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}
	return nil
}
