package format

// Config represents configuration options for formatters
type Config map[string]any

// GetString returns a string value from the config, or the default if not found
func (c Config) GetString(key string, defaultValue string) string {
	if val, ok := c[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetInt returns an int value from the config, or the default if not found
func (c Config) GetInt(key string, defaultValue int) int {
	if val, ok := c[key]; ok {
		// Handle both int and float64 (from JSON unmarshaling)
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// GetBool returns a bool value from the config, or the default if not found
func (c Config) GetBool(key string, defaultValue bool) bool {
	if val, ok := c[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetFloat returns a float64 value from the config, or the default if not found
func (c Config) GetFloat(key string, defaultValue float64) float64 {
	if val, ok := c[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultValue
}

// Has checks if a key exists in the config
func (c Config) Has(key string) bool {
	_, ok := c[key]
	return ok
}
