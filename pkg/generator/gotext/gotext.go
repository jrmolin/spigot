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
		"Percent": Percent,
		"PlusInt": PlusInt,
		"TimesInt": TimesInt,
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
		seconds := 0
		if trunc > 0 {
			seconds = rand.Intn(trunc)
		}
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

func ToInt(input any) int {

	if v, ok := input.(int); ok {
		return v
	}
	switch v := input.(type) {
	case string:
		result, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal("Could not convert %v (%T) to int: %v\n", input, input, err)
			return 1
		}
		return result
	default:
	}
	return 1

}

func PlusInt(a, b any) string {

	return fmt.Sprintf("%v", ToInt(a) + ToInt(b))
}
func TimesInt(a, b any) string {

	return fmt.Sprintf("%v", ToInt(a) * ToInt(b))
}

func Percent(numerator, denominator any) string {
	fnum := float64(ToInt(numerator))
	dnum := float64(ToInt(denominator))

	result := fnum / dnum * 100.0
	return fmt.Sprintf("%8.6f", result)
}

func RandomDuration() string {
	// return the string interpretation of that value
	return fmt.Sprintf("%01d:%02d:%02d", rand.Intn(4), rand.Intn(60), rand.Intn(60))
}

func RandomInt(maximum int) string {
	randval := 0
	if maximum > 0 {
		// get a random value
		randval = rand.Intn(maximum)
	} else if maximum < 0 {
		// get a random value
		randval = -1 * rand.Intn(-1 * maximum)
	}

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

type Field struct {
	Name string `config:"name"`
	Type string `config:"type"`
	Choices []string `config:"choices"`
	template *Template
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
		object[f.Name] = f.randomize(object)
	}

	// are there formats?
	if len(g.templates) < 1 {
		fmt.Printf("i am %v; %v\n", g.Name, g.templates)
		return nil, fmt.Errorf("This has no templates to process; bailing")
	}
	index := rand.Intn(len(g.templates))

	// attempt to generate each one
	err := g.templates[index].Tpl.Execute(&buf, object)
	if err != nil {
		log.Fatal("Failed to execute template", g.templates[index].Format, "with error", err)
		return nil, err
	}

	return buf.Bytes(), err
}

// New is Factory for the gotext generator
func New(cfg *ucfg.Config) (generator.Generator, error) {
	c := defaultConfig()
	if err := cfg.Unpack(&c); err != nil {
		return nil, err
	}

	gotextConfig := c.Config

	// check variables
	// return
	g := &GoText{
		Name: gotextConfig.Name,
		Fields: nil,
		templates: nil,
	}

	for i, v := range gotextConfig.Fields {
		f := Field{
			Name: v.Name,
			Type: v.Type,
			Choices: v.Choices,
		}

		// if there is a Template field
		if v.Template != nil {
			t, err := template.New(strconv.Itoa(i)).Funcs(FunctionMap).Parse(*v.Template)
			if err != nil {
				return nil, err
			}

			f.template = &Template{
				Format: *v.Template,
				Tpl: t,
			}
		}
		g.Fields = append(g.Fields, f)
	}

	for i, v := range gotextConfig.Formats {
		t, err := template.New(strconv.Itoa(i)).Funcs(FunctionMap).Parse(*v)
		if err != nil {
			return nil, err
		}
		g.templates = append(g.templates, Template{
			Format: *v,
			Tpl: t,
		})
	}

	return g, nil
}

