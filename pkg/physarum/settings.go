package physarum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Settings map[string]interface{}

var DefaultSettings = Settings{
	"width":       4096,
	"height":      2048,
	"initType":    "random_circle_random",
	"particles":   1 << 23,
	"fps":         60,
	"seed":        time.Now().UTC().UnixNano() - time.Date(2020, 3, 12, 9, 0, 0, 0, time.UTC).UnixNano(), // nanoseconds since psuedo-epoch
	"output_path": "output",
	"blurRadius":  1,
	"blurPasses":  2,
	"zoomFactor":  1,
	"scale":       0.5,
	"gamma":       1 / 2.2,
}

func NewSettings(settingsFile string) Settings {
	// TODO: Stub, replace with json read/write
	s := DefaultSettings
	rand.Seed(s["seed"].(int64))

	s["output_file"] = fmt.Sprint(s["seed"].(int64) / (1000 * 1000 * 1000)) // seconds since psuedo-epoch
	numConfigs := 2 + rand.Intn(4)
	s["configs"] = RandomConfigs(numConfigs)
	s["attract_table"] = RandomAttractionTable(numConfigs)
	s["palette"] = RandomPalette()

	if settingsFile != "" {
		fmt.Println(settingsFile)

		// Open our jsonFile, handle errors, and read the file
		jsonFile, err := os.Open(settingsFile)
		if err != nil {
			fmt.Println(err)
		}
		defer jsonFile.Close() // defer the closing of our jsonFile so that we can parse it later on
		jsonBytes, _ := ioutil.ReadAll(jsonFile)

		fmt.Print(string(jsonBytes))

		// Parse the json
		var data map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &data); err != nil {
			log.Println("Error parsing JSON string - ", err)
		}

		fmt.Println("*")
		fmt.Print(data)
		fmt.Println("*")
	}

	// panic("break")
	return s
	// return Settings{}
}

func (s Settings) GetSettingsJson() []byte {
	// Encode settings as json stored as an array of bytes
	b, err := json.Marshal(s)
	if err != nil {
		log.Fatalln(err)
	}

	return b
}

func (s Settings) GetOutputPath() string {
	// The file path to the output destination
	return s["output_path"].(string)
}

func (s Settings) GetFilePathWOExtension() string {
	// The file path and file base to the output destination without an extention
	return filepath.Join(s["output_path"].(string), s["output_file"].(string))
}

func (s Settings) WriteSettingsToFile() error {
	// Write a json file with the settings
	if err := os.MkdirAll(s.GetOutputPath(), os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	return ioutil.WriteFile(s.GetFilePathWOExtension()+".json", s.GetSettingsJson(), 0644)
}
