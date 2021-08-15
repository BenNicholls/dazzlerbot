package main

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

type configuration struct {
	Token             string   `yaml:"Api Token"`
	DefaultPrefixLen  int      `yaml:"Default Prefix Length"`
	SentenceLen       int      `yaml:"Default Sentence Max Length"`
	ResponseFrequency int      `yaml:"Random Response Frequency"`
	TriggerWords      []string `yaml:"Trigger Words"`
	//BlacklistIDs      []string    `yaml:"Blacklisted Users"`
}

//Generates a default config file and writes it to disk.
func writeDefaultConfig() error {
	defaultConfig := configuration{
		Token:             "PUT YOUR DISCORD API TOKEN HERE",
		DefaultPrefixLen:  2,
		SentenceLen:       100,
		ResponseFrequency: 15,
		TriggerWords:      make([]string, 0),
		//BlacklistIDs:      make([]string, 0),
	}

	f, ferr := os.Create("config.txt")
	if ferr != nil {
		return errors.New("Error creating config file: " + ferr.Error())
	}
	defer f.Close()
	data, derr := yaml.Marshal(&defaultConfig)
	if derr != nil {
		return errors.New("Error creating config file: " + derr.Error())
	}
	f.Write(data)

	return nil
}

//validates values of the configuration. returns nil if it's a-ok
func (c configuration) validate() error {
	if c.DefaultPrefixLen < 1 {
		return errors.New("Error importing config: Default Prefix Length cannot be less than 1.")
	}

	return nil
}
