package config

import (
	"os"

	"github.com/joho/godotenv"
)

func NewEnv() *Env {
	return &Env{}
}

type Env struct {
}

func (e *Env) LoadEnv() {
	godotenv.Load(".env")
}

func (e *Env) Get(key string, defaultValue ...string) string {
	val := os.Getenv(key)

	if val == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return val
}
