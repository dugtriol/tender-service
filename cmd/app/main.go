package main

import "tender-service/internal/app"

const (
	configPath = "config/config.yaml"
)

func main() {
	app.Run(configPath)
}
