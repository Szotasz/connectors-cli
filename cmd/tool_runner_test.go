package cmd

import (
	"reflect"
	"strings"
	"testing"
)

// Through v0.3.0 the object-typed flags (falai generate --params) were
// forwarded as raw strings, so the gateway validator always answered
// "Missing required parameter" for the first required field regardless of
// the flag's content. These pin the parse layer that closes that path.
func TestParseJSONArgObject(t *testing.T) {
	got, err := parseJSONArg("params", "object", `{"prompt":"a red fox in a forest"}`)
	if err != nil {
		t.Fatalf("valid object rejected: %v", err)
	}
	want := map[string]interface{}{"prompt": "a red fox in a forest"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parsed value mismatch: got %#v", got)
	}
}

func TestParseJSONArgInvalidJSONFailsLocallyAndNamesTheFlag(t *testing.T) {
	_, err := parseJSONArg("params", "object", `{"prompt":`)
	if err == nil {
		t.Fatal("invalid JSON accepted -- it would travel to the server as a bogus string")
	}
	if !strings.Contains(err.Error(), "--params") {
		t.Fatalf("error must name the flag for the user, got: %v", err)
	}
}

func TestParseJSONArgObjectRejectsNonObjectShape(t *testing.T) {
	if _, err := parseJSONArg("params", "object", `[1,2]`); err == nil {
		t.Fatal("array accepted for an object-typed flag")
	}
	if _, err := parseJSONArg("params", "object", `"just a string"`); err == nil {
		t.Fatal("quoted string accepted for an object-typed flag")
	}
}

func TestParseJSONArgArray(t *testing.T) {
	got, err := parseJSONArg("ids", "array", `[1,2,3]`)
	if err != nil {
		t.Fatalf("valid array rejected: %v", err)
	}
	if arr, ok := got.([]interface{}); !ok || len(arr) != 3 {
		t.Fatalf("parsed array mismatch: got %#v", got)
	}
	if _, err := parseJSONArg("ids", "array", `{"a":1}`); err == nil {
		t.Fatal("object accepted for an array-typed flag")
	}
}
