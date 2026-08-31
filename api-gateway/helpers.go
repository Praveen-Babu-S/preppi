package main

import (
	"fmt"
	"os"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func addr(port int) string {
	return fmt.Sprintf(":%d", port)
}
