package service

import (
	"encoding/json"
	"reflect"
	"regexp"
	"testing"
)

var (
	fieldNameRegexp = regexp.MustCompile("^[a-z0-9_]+$")
)

func TestEventMembersHaveCorrectJSONCase(t *testing.T) {
	events := []interface{}{
		ServiceStateChangedEvent{},
		ViewStateChangedEvent{},
	}

	for i, e := range events {
		// Serialize each event struct to JSON and back again
		// into a map so that we can inspect the resulting
		// names of the fields.
		buf, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("Failed to marshal message %d to JSON: %+v: %s", i, e, err)
		}
		var m map[string]interface{}
		err = json.Unmarshal(buf, &m)
		if err != nil {
			t.Fatalf("Failed to unmarshal message %d to JSON: %q: %s", i, string(buf), err)
		}
		checkMapKeyNames(m, t)
	}
}

func checkMapKeyNames(m map[string]interface{}, t *testing.T) {
	for k := range m {
		if !fieldNameRegexp.MatchString(k) {
			t.Errorf("JSON field name %q doesn't match the required pattern %q",
				k, fieldNameRegexp.String())
		}
		v := m[k]
		if reflect.ValueOf(v).Kind() == reflect.Map {
			checkMapKeyNames(v.(map[string]interface{}), t)
		}
	}
}
