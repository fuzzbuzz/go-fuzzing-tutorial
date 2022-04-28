package advancedtechniques

import (
	"reflect"
	"testing"

	yaml2 "github.com/goccy/go-yaml"
	yaml1 "gopkg.in/yaml.v2"
)

func FuzzYamlDifferential(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		map1 := map[string]interface{}{}
		map2 := map[string]interface{}{}

		err1 := yaml1.Unmarshal(data, &map1)
		err2 := yaml2.Unmarshal(data, &map2)

		if err1 == nil && err2 == nil {
			if len(map1) == 0 && len(map2) == 0 {
				// Reflect.DeepEqual doesn't handle this case well
				return
			}

			// If both think the data is valid, make sure they got the same structure
			if !reflect.DeepEqual(map1, map2) {
				t.Logf("Yaml1: %+v", map1)
				t.Logf("Yaml2: %+v", map2)
				t.Fatalf("Parsed yaml mismatch.")
			}
		}
	})
}
