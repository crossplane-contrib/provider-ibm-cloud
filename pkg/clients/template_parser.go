/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clients

import (
	"bytes"
	"encoding/json"
	"text/template"
)

// TemplateParser handles go templating for connection variables
type TemplateParser struct {
	Vars map[string]string
	Obj  map[string]interface{}
}

// NewTemplateParser returns a new TemplateParserParser for an object obj to parse and a set
// of variables and templates vars.
func NewTemplateParser(vars map[string]string, obj map[string]interface{}) *TemplateParser {
	return &TemplateParser{
		Vars: vars,
		Obj:  obj,
	}
}

// Parse parses the template, evaluates and store result
func (c *TemplateParser) Parse() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	for k, t := range c.Vars {
		tmpl, err := template.New("set").Funcs(template.FuncMap{"json": marshal}).Parse(t)
		if err != nil {
			return data, err
		}

		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, c.Obj)
		if err != nil {
			return data, err
		}

		data[k] = buf.String()
	}
	return data, nil
}

// marshal wraps json.Marshal to return string instead of []byte
func marshal(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
