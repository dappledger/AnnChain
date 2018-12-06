//json cfg file utility
package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	data map[string]interface{}
	Raw  []byte
}

func newConfig() *Config {
	result := new(Config)
	result.data = make(map[string]interface{})
	return result
}

// Loads config information from a JSON file
func LoadConfigFile(filename string) *Config {
	result := newConfig()
	err := result.parse(filename)
	if err != nil {
		log.Fatalf("error loading config file %s: %s", filename, err)
	}
	return result
}

// Loads config information from a JSON string
func LoadConfigString(s string) *Config {
	result := newConfig()
	err := json.Unmarshal([]byte(s), &result.data)
	if err != nil {
		log.Fatalf("error parsing config string %s: %s", s, err)
	}
	return result
}

func (c *Config) parse(fileName string) error {
	jsonFileBytes, err := ioutil.ReadFile(fileName)
	c.Raw = jsonFileBytes
	if err == nil {
		err = json.Unmarshal(jsonFileBytes, &c.data)
	}
	return err
}

// Returns a string for the config variable key
func (c *Config) GetString(key string) string {
	result, present := c.data[key]
	if !present {
		return ""
	}
	return result.(string)
}

// Returns a float for the config variable key
func (c *Config) GetFloat(key string) float64 {
	x, ok := c.data[key]
	if !ok {
		return -1
	}
	return x.(float64)
}

func (c *Config) GetInt(key string) int64 {
	return int64(c.GetFloat(key))
}

// Returns a bool for the config variable key
func (c *Config) GetBool(key string) bool {
	x, ok := c.data[key]
	if !ok {
		return false
	}
	return x.(bool)
}

// Returns an array for the config variable key
func (c *Config) GetArray(key string) []interface{} {
	result, present := c.data[key]
	if !present {
		return []interface{}(nil)
	}
	return result.([]interface{})
}
