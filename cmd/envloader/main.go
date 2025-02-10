package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// getEnvAsMap retrieves all environment variables as a map.
func getEnvAsMap() map[string]string {
	envs := make(map[string]string)
	for _, env := range os.Environ() {
		pair := splitEnv(env)
		envs[pair[0]] = pair[1]
	}

	return envs
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

func main() {
	envs := getEnvAsMap()

	envJSON, err := json.Marshal(envs)
	if err != nil {
		log.Error().Err(err).Msg("Error marshalling envs to JSON")

		return
	}

	//nolint:forbidigo
	fmt.Println(string(envJSON))
}
