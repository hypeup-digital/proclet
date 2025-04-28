package config

import "time"

type Application struct {
	Identifier string
	Command    string
}

type OutputConfig struct {
	PrintAppNames    bool
	PrintTimeStamps  bool
	MaxAppNameLength int
}

type Config struct {
	Banner       string
	Applications []Application
	Timeout      time.Duration
	Output       OutputConfig
}
