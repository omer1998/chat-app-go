package app

import (
	"bufio"
	"encoding/json"
	"errors"
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
	BasePath string
}

const configFileName = "config.json"

func NewConfig() (*Config, error) {
	basePath := "C:/Users/master/Desktop/chat-app-go/chat/zarf"
	path := filepath.Join(basePath, configFileName)
	if _, err := os.Stat(path); err != nil {
		f, err := os.Create(path)
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
			BasePath: basePath,
		}, nil

	}
	f, err := os.Open(path)
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
		BasePath: basePath,
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
		BasePath: c.BasePath,
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
	filePath := filepath.Join(c.BasePath, id+".msg")
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// create this file
			f, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("error creating msg file: %w", err)
			}
			defer f.Close()
			f.WriteString(msg)
			return nil
		}
		return fmt.Errorf("error stat msgs file: %w", err)
	}
	f, err := os.OpenFile(filePath, os.O_APPEND, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error opening msgs file: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(msg)
	if err != nil {
		return fmt.Errorf("error writing msg to file: %w", err)
	}

	return nil
}
func (c *Config) GetMsgsFromFile(id string) []string {
	// the file name is constructed like this; id.msg
	filePath := filepath.Join(c.BasePath, id+".msg")
	f, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return []string{}
	}
	defer f.Close()
	var messages []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		messages = append(messages, scanner.Text())
	}
	// if errors.Is(scanner.Err(), io.EOF) {
	// 	return messages
	// } else if scanner.Err() != nil {
	// 	return []string{}

	// }
	return messages

}

// ====================================================================================

func writeConfig(cnfg Config) error {
	filePath := filepath.Join(cnfg.BasePath, configFileName)
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("writeConfig, error create/truncate config file: %w \n path: %s", err, cnfg.BasePath)
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
