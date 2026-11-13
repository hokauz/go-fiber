package envloader

import (
	"fmt"
	"os"
)

// Message levels.
const (
	LevelError   = "error"
	LevelWarning = "warning"
)

// Error kinds.
const (
	KindMissing = "missing"
	KindEmpty   = "empty"
	KindInvalid = "invalid"
)

// Value types for EnvDef.
const (
	TypeString = "string"
	TypeInt    = "int"
	TypeBool   = "bool"
)

// EnvMessage represents a validation message with level (error or warning).
type EnvMessage struct {
	Level   string
	Key     string
	Kind    string
	Message string
}

func (m EnvMessage) String() string {
	return fmt.Sprintf("%s [%s]: %s", m.Key, m.Kind, m.Message)
}

// EnvDef describes an env variable: name, type, required, default.
// Orients the Load flow for each entry.
// Default is used when optional and missing or empty.
type EnvDef struct {
	Name     string
	Type     string
	Required bool
	Default  any
}

// Envs holds configuration loaded from environment variables.
type Envs struct {
	Values map[string]any
}

// DefaultDefs is the predetermined list for Load.
// Empty by default; other modules pass their own defs via LoadWith.
var DefaultDefs = []EnvDef{}

// Exists returns true if the env key is set. Uses LookupEnv (before Getenv).
func Exists(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

// Read returns the value of the env. Call after Exists.
func Read(key string) string {
	return os.Getenv(key)
}

// Load executes the flow for DefaultDefs. Returns (nil, msgs) when required fails.
func Load() (*Envs, []EnvMessage) {
	return LoadWith(DefaultDefs)
}

// LoadWith executes the flow for the given list.
func LoadWith(defs []EnvDef) (*Envs, []EnvMessage) {
	var msgs []EnvMessage
	var requiredFailed bool
	cfg := &Envs{Values: make(map[string]any)}

	for _, def := range defs {
		// 1. Exists?
		if !Exists(def.Name) {
			if def.Required {
				requiredFailed = true
				msgs = append(msgs, EnvMessage{Level: LevelError, Key: def.Name, Kind: KindMissing, Message: def.Name + " missing"})
			} else if def.Default != nil {
				cfg.Values[def.Name] = def.Default
			}
			continue
		}

		// 2. Read
		v := Read(def.Name)

		// 3. Format OK?
		parsed, ok, msg := validate(def, v)
		if !ok {
			if !def.Required && def.Default != nil {
				cfg.Values[def.Name] = def.Default
				msgs = append(msgs, EnvMessage{Level: LevelWarning, Key: def.Name, Kind: KindInvalid, Message: msg})
			} else {
				kind := KindInvalid
				if v == "" {
					kind = KindEmpty
				}
				msgs = append(msgs, EnvMessage{Level: LevelError, Key: def.Name, Kind: kind, Message: msg})
				if def.Required {
					requiredFailed = true
				}
			}
			continue
		}

		// 4. Insert
		cfg.Values[def.Name] = parsed
	}

	if requiredFailed {
		return nil, msgs
	}
	return cfg, msgs
}

// GetString returns the string value for key. Panics if missing or wrong type.
func (e *Envs) GetString(key string) string {
	return e.Values[key].(string)
}

// GetInt returns the int value for key. Panics if missing or wrong type.
func (e *Envs) GetInt(key string) int {
	return e.Values[key].(int)
}

// GetBool returns the bool value for key. Panics if missing or wrong type.
func (e *Envs) GetBool(key string) bool {
	return e.Values[key].(bool)
}
