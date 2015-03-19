package main

import (
	"config"
	"export"
	"flag"
	"fmt"
	"openweathermap"
	"strconv"
	"time"

	mylog "github.com/patrickalin/GoMyLog"
)

/*
Get Nest Thermostat Information
*/

//name of the config file
const configName = "config"

var (
	openWeathermapMessageToInfluxDB = make(chan openweathermap.OpenweatherStruct)

	myTime time.Duration

	myConfig config.ConfigStructure

	debug = flag.String("debug", "", "Error=1, Warning=2, Info=3, Trace=4")
)

func main() {

	flag.Parse()

	fmt.Printf("\n %s :> OpenWeather Go Call to InfluxDB\n\n", time.Now().Format(time.RFC850))

	mylog.Init(mylog.ERROR)

	// getConfig from the file config.json
	myConfig = config.New(configName)

	if *debug != "" {
		myConfig.LogLevel = *debug
	}

	level, _ := strconv.Atoi(myConfig.LogLevel)
	mylog.Init(mylog.Level(level))

	i, _ := strconv.Atoi(myConfig.RefreshTimer)
	myTime = time.Duration(i) * time.Second

	//init listeners
	if myConfig.InfluxDBActivated == "true" {
		export.InitInfluxDB(openWeathermapMessageToInfluxDB, myConfig)
	}

	schedule()
}

func schedule() {
	ticker := time.NewTicker(myTime)
	quit := make(chan struct{})
	repeat()
	for {
		select {
		case <-ticker.C:
			repeat()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func repeat() {

	go func() {
		myOpenWeathermap, err := openweathermap.MakeNew(myConfig)
		if err == nil {
			openWeathermapMessageToInfluxDB <- myOpenWeathermap
		}
	}()

}
