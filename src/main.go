// How to implement Boiler Plate structure for an api in GIN in real world.
// https://github.com/RohitBinjola/GolangAPIBoilerPlate/blob/main/Demo/main.go

package main

import (
	"challenge/config"
	"challenge/database/database"
	"challenge/logger"
	"challenge/router"
)

func main() {
	config.Init()
	config.Appconfig = config.GetConfig()
	logger.Init()
	logger.InfoLn("Logger Initialized successfully")
	database.Init()
	logger.InfoLn("Database Initialize successfully")
	router.Init()
	logger.InfoLn("Router Initialized successfully")
}
