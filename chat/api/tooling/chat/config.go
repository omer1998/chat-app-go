package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
)

var errUserNotFound = fmt.Errorf("user not found")

// the whole idea of this file is to persist data of user his contact in a specific json file under zarf
type user struct {
	Id       string
	Name     string
	Messages []string
}

type document struct {
	User     user   `json:"user"`
	Contacts []user `json:"contacts"`
}

type Config struct {
	User     user
	Contacts []user
	FilePath string
}

func NewConfig() (*Config, error) {
	filePath := filepath.Join("C:/Users/master/Desktop/chat-app-go/chat/zarf", "config.json")
	if _, err := os.Stat(filePath); err != nil {
		f, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("error creating config file: %w", err)
		}
		defer f.Close()
		doc := document{
			User: user{
				Id:   strconv.Itoa(rand.Intn(1000000)),
				Name: "omer faris",
			},
			// Users:  at the beggining is empty for this user
		}
		// we need to write this data to the file
		data, err := json.MarshalIndent(doc, " ", "	")
		if err != nil {
			return nil, fmt.Errorf("error marshaling json data: %w", err)
		}
		if _, err := f.Write(data); err != nil {
			return nil, fmt.Errorf("error writing json data to file: %w", err)
		}
		return &Config{
			User:     doc.User,
			FilePath: filePath,
		}, nil

	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error openning file: %w", err)
	}
	defer f.Close()

	var doc document
	if err := json.NewDecoder(f).Decode(&doc); err != nil {
		return nil, fmt.Errorf("error decoding file: %w", err)

	}
	return &Config{
		User:     doc.User,
		Contacts: doc.Contacts,
		FilePath: filePath,
	}, nil
}

// func (c *Config) AddContact(usr user) error {}
// Lookup function is used to look if this user present in our contact list
func (c *Config) LookUpUser(id string) (user, error) {
	for i, usr := range c.Contacts {
		if usr.Id == id {
			return c.Contacts[i], nil
		}
	}
	return user{}, errUserNotFound
}

func (c *Config) UpdateContact(usr user) error {
	c.Contacts = append(c.Contacts, usr)
	if err := writeConfig(Config{
		User:     c.User,
		Contacts: c.Contacts,
		FilePath: c.FilePath,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Config) AddMessage(id string, msg string) error {
	for i, usr := range c.Contacts {
		if id == c.Contacts[i].Id {
			usr.Messages = append(usr.Messages, msg)
			c.Contacts[i] = usr
		}
	}
	return nil
}

// ====================================================================================

func writeConfig(cnfg Config) error {
	f, err := os.Create(cnfg.FilePath)
	if err != nil {
		return fmt.Errorf("writeConfig, error create/truncate config file: %w \n path: %s", err, cnfg.FilePath)
	}
	defer f.Close()
	doc := document{
		User:     cnfg.User,
		Contacts: cnfg.Contacts,
	}
	data, err := json.MarshalIndent(doc, "", "	")
	if err != nil {
		return fmt.Errorf("write config, error marshal document: %w", err)
	}
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("error writing data to config file: %w ", err)
	}
	return nil

}
