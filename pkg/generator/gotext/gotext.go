// Package gotext implements the generator for generic logs.
//
// Configuration file supports including timestamps in log messages
//
//   generator:
//     type: gotext
//     include_timestamp: true
package gotext

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/elastic/go-ucfg"
	"github.com/elastic/spigot/pkg/generator"
	"github.com/elastic/spigot/pkg/random"
)

// Name is the name of the generator in the configuration file and registry
const Name = "gotext"

func init() {
	generator.Register(Name, New)
}
var (
	FunctionMap = template.FuncMap{
		"ToLower": strings.ToLower,
		"ToUpper": strings.ToUpper,
		"TimestampFormatter": TimestampFormatter,
		"RandomIPv4": RandomIPv4,
		"RandomPort": RandomPort,
		"RandomInt": RandomInt,
	}
)
func TimestampFormatter(format, whence string) string {
	now := time.Now()

	dur,err := time.ParseDuration(whence)
	if err != nil {
		fmt.Println("failed to parse [", whence, "] with error::", err)
	} else {
		// now choose a random duration within this value
		trunc := int(dur.Round(time.Second).Seconds())
		seconds := rand.Intn(trunc)
		newdur, err := time.ParseDuration(fmt.Sprintf("-%ds", seconds))
		if err != nil {
			// ignore
		} else {
			dur = newdur
		}
		now = now.Add(dur)
	}

	// this format is seconds.[1-9]
	if strings.HasPrefix(format, "seconds") {
		secs := now.Unix()

		returnVal := fmt.Sprintf("%d", secs)
		result := strings.Split(format, ".")
		if len(result) > 1 {
			// generate a random number of nanos
			nanos := rand.Intn(1_000_000_000)
			partialsString := fmt.Sprintf("%09d", nanos)

			// figure out what precision to output
			precision := 6
			if prec, err := strconv.Atoi(result[1]); err == nil {
				switch prec {
				case 0:
					return fmt.Sprintf("%d", secs)
				case 1, 2, 3, 4, 5, 6, 7, 8, 9:
					precision = prec
				default:
					precision = 9
				}
			}
			returnVal = fmt.Sprintf("%d.%s", secs, partialsString[0:precision])
		}
		// return early
		return returnVal
	}

	return now.Format(format)
}

func RandomDuration() string {
	// return the string interpretation of that value
	return fmt.Sprintf("%01d:%02d:%02d", rand.Intn(4), rand.Intn(60), rand.Intn(60))
}

func RandomInt(maximum int) string {
	// get a random value
	randval := rand.Intn(maximum)

	// return the string interpretation of that value
	return strconv.Itoa(randval)
}

func RandomIPv4() string {
	return random.IPv4().String()
}

func RandomPort() string {
	return strconv.Itoa(random.Port())
}

type Template struct {
	Format string
	Tpl *template.Template
}

type GoText struct {
	Name string
	Fields []Field
	templates []Template
}

func (g *GoText) Next() ([]byte, error) {
	var buf bytes.Buffer

	object := make(map[string]any)

	object["Timestamp"] = time.Now()

	// loop over each field
	for _, f := range g.Fields {
		object[f.Name] = f.randomize()
	}

	// are there formats?
	index := rand.Intn(len(g.templates))

	// attempt to generate each one
	err := g.templates[index].Tpl.Execute(&buf, object)
	if err != nil {
		log.Fatal("Failed to execute template", g.templates[index].Format, "with error", err)
		return nil, err
	}

	return buf.Bytes(), err
}

// New is Factory for the asa generator
func New(cfg *ucfg.Config) (generator.Generator, error) {
	c := defaultConfig()
	if err := cfg.Unpack(&c); err != nil {
		return nil, err
	}

	gc := c.Config

	// check variables
	// return
	g := &GoText{
		Name: gc.Name,
		Fields: gc.Fields,
		templates: nil,
	}

	for i, v := range gc.Formats {
		t, err := template.New(strconv.Itoa(i)).Funcs(FunctionMap).Parse(v.Value)
		if err != nil {
			return nil, err
		}
		g.templates = append(g.templates, Template{
			Format: v.Value,
			Tpl: t,
		})
	}

	return g, nil
}

