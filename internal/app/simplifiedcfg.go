package app

import (
	"fmt"
	"os"
	"strings"
)

// ResolveSimplified returns whether tools default to simplified Chinese. flag is nil when --simplified
// was not given; otherwise it wins. Without it FHL_SIMPLIFIED (1/true/yes/on, 0/false/no/off) applies,
// then false.
func ResolveSimplified(flag *bool) (bool, error) {
	if flag != nil {
		return *flag, nil
	}
	v := strings.ToLower(strings.TrimSpace(os.Getenv(EnvSimplified)))
	switch v {
	case "":
		return false, nil
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	}
	return false, fmt.Errorf("%s=%q must be 1/true/yes/on or 0/false/no/off", EnvSimplified, os.Getenv(EnvSimplified))
}
