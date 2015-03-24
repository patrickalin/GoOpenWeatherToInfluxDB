package export

import (
	"config"
	"encoding/json"
	"fmt"
	"openweathermap"
	"time"

	mylog "github.com/patrickalin/GoMyLog"
	rest "github.com/patrickalin/GoRest"
)

type influxDBStruct struct {
	Columns []string        `json:"columns"`
	Serie   string          `json:"name"`
	Points  [][]interface{} `json:"points"`
}

type influxDBError struct {
	message error
	advice  string
}

func (e *influxDBError) Error() string {
	return fmt.Sprintf("\n \t InfluxDBError :> %s \n\t InfluxDB Advice:> %s %s %s %s", e.message, e.advice)
}

func sendOpenWeatherToInfluxDB(oneOpenWeather openweathermap.OpenweatherStruct, oneConfig config.ConfigStructure) {

	fmt.Printf("\n %s :> Send OpenWeather Data to InfluxDB\n", time.Now().Format(time.RFC850))

	influxDBData := influxDBStruct{}

	//init colomn Name
	influxDBData.Columns = make([]string, 8)
	dataColumls := [8]string{"City", "Humidity", "Pressure", "Temperature", "WindSpeed", "WindDegree", "Sunrise", "Sunset"}
	for i := range dataColumls {
		influxDBData.Columns[i] = dataColumls[i]
	}

	//init array
	influxDBData.Points = make([][]interface{}, 1)
	for i := range influxDBData.Points {
		influxDBData.Points[i] = make([]interface{}, 8)
	}
	for i := range influxDBData.Points[0] {
		influxDBData.Points[0][i] = 0.0
	}

	influxDBData.Serie = "OpenWeather"

	influxDBData.Points[0][0] = oneOpenWeather.GetCity()
	influxDBData.Points[0][1] = oneOpenWeather.GetHumidity()
	influxDBData.Points[0][2] = oneOpenWeather.GetPressure()
	influxDBData.Points[0][3] = oneOpenWeather.GetTemp()
	influxDBData.Points[0][4] = oneOpenWeather.GetWindSpeed()
	influxDBData.Points[0][5] = oneOpenWeather.GetWindDeg()
	influxDBData.Points[0][6] = oneOpenWeather.GetSunrise()
	influxDBData.Points[0][7] = oneOpenWeather.GetSunset()
	//influxDBData.Points[0][8] = oneOpenWeather.GetDescription()

	err := sendPost(influxDBData, oneConfig)
	if err != nil {
		mylog.Error.Fatal(&influxDBError{err, "Error sent Data to Influx DB"})
	}

}

func sendPost(influxDBData interface{}, oneConfig config.ConfigStructure) (err error) {
	data, _ := json.Marshal(influxDBData)

	data = append(data, byte(']'))
	data = append([]byte(`[`), data...)

	fullURL := fmt.Sprint("http://", oneConfig.InfluxDBServer, ":", oneConfig.InfluxDBServerPort, "/db/", oneConfig.InfluxDBDatabase, "/series?u=", oneConfig.InfluxDBUsername, "&p=", oneConfig.InfluxDBPassword)

	//curl -X POST -d '[{"name":"foo","columns":["val"],"points":[[23]]}]' 'http://localhost:8086/db/nest/series?u=root&p=root'
	oneRest := rest.MakeNew()
	err = oneRest.PostJSON(fullURL, data)
	if err != nil {
		err2 := createDB(oneConfig)
		if err2 != nil {
			return &influxDBError{err2, "Error with Post : Check if InfluxDB is running or if the database nest exists"}
		}
	}
	return nil
}

func createDB(oneConfig config.ConfigStructure) error {
	type createDB struct {
		Name string `json:"name"`
	}

	fmt.Println("\n Create Database OpenWeather\n")

	nestDB := createDB{}
	fullURL := fmt.Sprint("http://", oneConfig.InfluxDBServer, ":", oneConfig.InfluxDBServerPort, "/db?u=", oneConfig.InfluxDBUsername, "&p=", oneConfig.InfluxDBPassword)
	nestDB.Name = oneConfig.InfluxDBDatabase
	data, _ := json.Marshal(nestDB)

	oneRest := rest.MakeNew()
	err := oneRest.PostJSON(fullURL, data)
	if err != nil {
		return &influxDBError{err, "Error with Post : create database OpenWeather"}
	}
	return nil
}

func InitInfluxDB(messagesOpenWeather chan openweathermap.OpenweatherStruct, oneConfig config.ConfigStructure) {

	go func() {
		mylog.Trace.Println("receive messagesOpenWeather  to export InfluxDB")
		for {
			msg := <-messagesOpenWeather
			sendOpenWeatherToInfluxDB(msg, oneConfig)
		}
	}()
}
