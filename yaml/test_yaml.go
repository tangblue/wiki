package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

var data_orig = `
a: 2018-04-21 22:14:43.256294
b:
  c: 2
  d: [d, e]
`
var output_file = "example.yaml"

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.
type T struct {
	A time.Time
	B struct {
		RenamedC int      `yaml:"c"`
		D        []string `yaml:",flow"`
	}
}

func main() {
	t := T{}

	data, err := ioutil.ReadFile(output_file)
	if err != nil {
		data = []byte(data_orig)
	}
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", t)

	d, err := yaml.Marshal(&t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)

	d, err = yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))

	err = ioutil.WriteFile(output_file, d, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
