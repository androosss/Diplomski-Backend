// Package dto (repo) contains data transfer objects (DTOs) used in repo package
package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

type Endpoint string

type Authentication struct {
	Endpoints []Endpoint    `json:"endpoints,omitempty"`
	Headers   UntypedConfig `json:"headers,omitempty"`
}

type Connection struct {
	Endpoints []Endpoint `json:"endpoints,omitempty"`
	Url       string     `json:"url,omitempty"`
}

type SubServer string

const (
	SUB_CL SubServer = "subServerForPublicClients"
	SUB_BO SubServer = "subServerForBackoffice"
	SUB_LO SubServer = "subServerForLogin"
)

type ConnectionSlice []Connection

func (cc ConnectionSlice) GetConnectionByEndpoint(endpoint Endpoint) Connection {
	for _, connection := range cc {
		for _, connectionEndpoint := range connection.Endpoints {
			if connectionEndpoint == endpoint {
				return connection
			}
		}
	}
	return Connection{}
}

type AuthenticationSlice []Authentication

type UntypedConfig map[string]interface{}

func (uc UntypedConfig) Value() (driver.Value, error) {
	return json.Marshal(uc)
}

func (uc *UntypedConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &uc)
}

type CommonConfig struct {
}

func (uc CommonConfig) Value() (driver.Value, error) {
	return json.Marshal(uc)
}

func (uc *CommonConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &uc)
}

func KeyToName(key string) string {
	if key == "PAYPAL" {
		return "PayPal"
	}
	name := ""
	for i, char := range key {
		if i == 0 {
			name += strings.ToUpper(string(char))
		} else {
			if char == '_' {
				name += string(' ')
			} else {
				name += strings.ToLower(string(char))
			}
		}
	}
	return name
}
