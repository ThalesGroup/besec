// Copyright (c) 2015-2016 Michael Persson
// Copyright (c) 2012–2015 Elasticsearch <http://www.elastic.co>
//
// Originally distributed as part of "beats" repository (https://github.com/elastic/beats).
// Modified specifically for "iodatafmt" package.
// via https://github.com/go-yaml/yaml/issues/139#issuecomment-220072190
//
// Distributed underneath "Apache License, Version 2.0" which is compatible with the LICENSE for this package.

package lib

import (
	// Base packages.
	"fmt"

	// Third party packages.
	"gopkg.in/yaml.v2"
)

// unmarshalYaml to map[string]interface{} instead of map[interface{}]interface{}.
func unmarshalYaml(in []byte, out interface{}) error {
	var res interface{}

	if err := yaml.Unmarshal(in, &res); err != nil {
		return err
	}
	*out.(*interface{}) = cleanupMapValue(res)

	return nil
}

func cleanupInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = cleanupMapValue(v)
	}
	return res
}

func cleanupInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	default:
		return v // in the original this was fmt.Sprintf("%v", v), which forced all values to strings. Not clear why.
	}
}
