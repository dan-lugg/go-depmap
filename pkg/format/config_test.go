package format

import "testing"

func TestConfig_GetString(t *testing.T) {
	config := Config{
		"key1": "value1",
		"key2": 123,
	}

	if got := config.GetString("key1", "default"); got != "value1" {
		t.Errorf("GetString() = %v, want %v", got, "value1")
	}

	if got := config.GetString("key2", "default"); got != "default" {
		t.Errorf("GetString() = %v, want %v", got, "default")
	}

	if got := config.GetString("missing", "default"); got != "default" {
		t.Errorf("GetString() = %v, want %v", got, "default")
	}
}

func TestConfig_GetInt(t *testing.T) {
	config := Config{
		"int":    42,
		"float":  3.14,
		"string": "not a number",
	}

	if got := config.GetInt("int", 0); got != 42 {
		t.Errorf("GetInt() = %v, want %v", got, 42)
	}

	if got := config.GetInt("float", 0); got != 3 {
		t.Errorf("GetInt() = %v, want %v", got, 3)
	}

	if got := config.GetInt("string", 99); got != 99 {
		t.Errorf("GetInt() = %v, want %v", got, 99)
	}

	if got := config.GetInt("missing", 100); got != 100 {
		t.Errorf("GetInt() = %v, want %v", got, 100)
	}
}

func TestConfig_GetBool(t *testing.T) {
	config := Config{
		"true":   true,
		"false":  false,
		"string": "not a bool",
	}

	if got := config.GetBool("true", false); got != true {
		t.Errorf("GetBool() = %v, want %v", got, true)
	}

	if got := config.GetBool("false", true); got != false {
		t.Errorf("GetBool() = %v, want %v", got, false)
	}

	if got := config.GetBool("string", true); got != true {
		t.Errorf("GetBool() = %v, want %v", got, true)
	}

	if got := config.GetBool("missing", false); got != false {
		t.Errorf("GetBool() = %v, want %v", got, false)
	}
}

func TestConfig_GetFloat(t *testing.T) {
	config := Config{
		"float":  3.14,
		"string": "not a number",
	}

	if got := config.GetFloat("float", 0.0); got != 3.14 {
		t.Errorf("GetFloat() = %v, want %v", got, 3.14)
	}

	if got := config.GetFloat("string", 1.5); got != 1.5 {
		t.Errorf("GetFloat() = %v, want %v", got, 1.5)
	}

	if got := config.GetFloat("missing", 2.5); got != 2.5 {
		t.Errorf("GetFloat() = %v, want %v", got, 2.5)
	}
}

func TestConfig_Has(t *testing.T) {
	config := Config{
		"key1": "value",
		"key2": nil,
	}

	if !config.Has("key1") {
		t.Error("Has() = false, want true for existing key")
	}

	if !config.Has("key2") {
		t.Error("Has() = false, want true for key with nil value")
	}

	if config.Has("missing") {
		t.Error("Has() = true, want false for missing key")
	}
}
