package aso

import (
	"flag"
	"strings"
)

// resolveArgs re-scans args to handle flags interspersed with positional
// arguments, which Go's flag package doesn't support natively. It updates
// the provided FlagSet with any flag values found after positional args
// and returns only the non-flag positional arguments.
//
// Comma-delimited values in positional args are expanded when expand is true,
// e.g. "photo editor,camera,vpn" → ["photo editor", "camera", "vpn"].
func resolveArgs(fs *flag.FlagSet, args []string, expand bool) []string {
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle --flag=value or -flag=value.
		if strings.HasPrefix(arg, "-") {
			name := strings.TrimLeft(arg, "-")
			if idx := strings.Index(name, "="); idx >= 0 {
				flagName := name[:idx]
				if f := fs.Lookup(flagName); f != nil {
					_ = f.Value.Set(name[idx+1:])
					continue
				}
			} else if f := fs.Lookup(name); f != nil {
				// --flag value (next arg is the value).
				if i+1 < len(args) {
					_ = f.Value.Set(args[i+1])
					i++
					continue
				}
			}
		}

		// Positional arg — optionally expand commas.
		if expand {
			for _, part := range strings.Split(arg, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					positional = append(positional, part)
				}
			}
		} else {
			positional = append(positional, arg)
		}
	}
	return positional
}
