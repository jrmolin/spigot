package gotext

import (
	"fmt"
	"math/rand"

	"strconv"

	"github.com/elastic/spigot/pkg/random"
)

type config struct {
	Type string `config:"type" validate:"required"`
	File string `config:"file" validate:"required"`
	Subtype string `config:"subtype" validate:"required"`
}

type format struct {
	Id string `config:"id" validate:"required"`
	Value string `config:"value" validate:"required"`
}

type Field struct {
	Name string `config:"name"`
	Type string `config:"type"`
	Choices []string `config:"choices"`
}

type generator_config struct {
	Name string `config:"name" validate:"required"`
	IncludeTimestamp bool `config:"include_timestamp"`
	Formats []*format `config:"formats"`
	Fields []Field `config:"fields"`
}

func defaultConfig() config {
	return config{
		Type: Name,
		File: "/does/not/exist.yml",
	}
}

func (c *config) Validate() error {
	if c.Type != Name {
		return fmt.Errorf("'%s' is not a valid value for 'type' expected '%s'", c.Type, Name)
	}

	return nil
}

func (f *Field) randomize() any {
	// if there are choices, select one at random
	if f.Choices != nil {
		return f.Choices[rand.Intn(len(f.Choices))]
	}

	// if there is a random definition, use that
	switch f.Type {
	case "IPv4", "IP", "ipv4":
		return RandomIPv4()
	case "Port", "port":
		return strconv.Itoa(random.Port())
	case "interface", "Intf", "intf":
		return fmt.Sprintf("%s%02d", f.Name, rand.Intn(16))
	case "Duration", "duration":
		return RandomDuration()
	default:
	}

	// otherwise, return the type as a string
	return f.Type

}
