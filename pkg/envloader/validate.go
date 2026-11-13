package envloader

import (
	"fmt"
	"strconv"
)

// validate runs format check for def+value. Returns (parsed, ok, errMsg).
func validate(def EnvDef, value string) (any, bool, string) {
	switch def.Type {
	case TypeString:
		if value == "" {
			return "", false, def.Name + " empty"
		}
		return value, true, ""
	case TypeInt:
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, false, fmt.Sprintf("%s=%q expected integer", def.Name, value)
		}
		return n, true, ""
	case TypeBool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, false, fmt.Sprintf("%s=%q expected boolean", def.Name, value)
		}
		return b, true, ""
	default:
		return nil, false, fmt.Sprintf("%s: unknown type %q", def.Name, def.Type)
	}
}
