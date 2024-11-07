package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	// Get all environment variables
	envs := make(map[string]string)
	for _, env := range os.Environ() {
		// Each env is in "KEY=VALUE" format
		pair := splitEnv(env)
		envs[pair[0]] = pair[1]
	}

	// Convert the map to a JSON string
	envJSON, err := json.Marshal(envs)
	if err != nil {
		log.Error().Err(err).Msg("Error marshalling envs to JSON")

		return
	}

	// Output the JSON string
	//nolint:forbidigo
	fmt.Println(string(envJSON))
}

// splitEnv splits an environment variable into key and value parts.
func splitEnv(env string) []string {
	for i := range env {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}

	return []string{env, ""}
}
