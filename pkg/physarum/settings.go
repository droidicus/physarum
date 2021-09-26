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

type Settings struct {
	// Not Exported!
	outputFile string
	outputPath string

	// Exported below this line
	Width         int
	Height        int
	Particles     int
	Fps           int
	StepsPerFrame int
	Seed          int64
	NumConfigs    int
	BlurRadius    int
	BlurPasses    int
	ZoomFactor    float32
	Scale         float32
	Gamma         float32
	InitType      string
	SaveVideo     bool

	AttractionTable [][]float32
	Configs         []Config
	Palette         Palette

	// Misc map[string]interface{}
}

func nsSincePsuedoEpoch() int64 {
	psuedo_epoch := time.Date(2020, 3, 12, 9, 0, 0, 0, time.UTC).UnixNano()
	return time.Now().UTC().UnixNano() - psuedo_epoch // nanoseconds since psuedo-epoch
}

func NewSettings(inputSettingsFile string) *Settings {
	// Simple Defaults
	s := &Settings{
		Width:         4096,
		Height:        2048,
		Particles:     1 << 23,
		Fps:           60,
		StepsPerFrame: 1,
		Seed:          nsSincePsuedoEpoch(),
		InitType:      "random_circle_random",
		BlurRadius:    1,
		BlurPasses:    2,
		ZoomFactor:    1,
		Scale:         0.5,
		Gamma:         1 / 2.2,
		outputPath:    "output",
		SaveVideo:     true,
	}

	// Read the JSON file settings if supplied, use default values (above) for fields not found
	if inputSettingsFile != "" {
		err := s.ReadSettingsFromFile(inputSettingsFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Set the seed according to the settings for deterministic configuration generation
	rand.Seed(s.Seed)

	// seconds since psuedo-epoch
	s.outputFile = fmt.Sprint(nsSincePsuedoEpoch() / (1000 * 1000 * 1000))

	// If NumConfigs is not specified, random palette
	if s.Palette == nil {
		s.Palette = RandomPalette()
	}

	// If NumConfigs is not specified, random value (note, this is not used unless the fields below are nil)
	if s.NumConfigs == 0 {
		s.NumConfigs = 2 + rand.Intn(4)
	}

	// If Configs is not specified, random config
	if s.Configs == nil {
		s.Configs = RandomConfigs(s.NumConfigs)
	}

	// If AttractionTable is not specified, random attraction table
	if s.AttractionTable == nil {
		s.AttractionTable = RandomAttractionTable(s.NumConfigs)
	}

	return s
}

func (s Settings) GetSettingsJson() []byte {
	// Encode settings as json stored as an array of bytes
	json_bytes, err := json.Marshal(s)
	if err != nil {
		log.Fatalln(err)
	}

	return json_bytes
}

func (s Settings) GetOutputPath() string {
	// The file path to the output destination
	return s.outputPath
}

func (s Settings) GetFilePathWOExtension() string {
	// The file path and file base to the output destination without an extention
	return filepath.Join(s.outputPath, s.outputFile)
}

func (s Settings) WriteSettingsToFile() error {
	// Write a json file with the settings
	if err := os.MkdirAll(s.GetOutputPath(), os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	// Write file
	return ioutil.WriteFile(s.GetFilePathWOExtension()+".json", s.GetSettingsJson(), 0644)
}

func (s *Settings) ReadSettingsFromFile(inputSettingsFile string) error {
	// Open our jsonFile, handle errors, and read the file
	jsonFile, err := os.Open(inputSettingsFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer jsonFile.Close() // defer the closing of our jsonFile so that we can parse it later on

	// Read the entire file into memory
	jsonBytes, _ := ioutil.ReadAll(jsonFile)

	// This will read the json file, and overwrite values from the current settings object with the fields found
	// Return error if any if returned
	return json.Unmarshal(jsonBytes, s)
}
