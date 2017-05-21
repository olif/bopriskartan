package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

type dataConfig struct {
	Name          string  `yaml:"name"`
	HeatmapWidth  int     `yaml:"heatmap_width"`
	HeatmapHeight int     `yaml:"heatmap_height"`
	LatLo         float64 `yaml:"lat_lo"`
	LonLo         float64 `yaml:"lng_lo"`
	LatHi         float64 `yaml:"lat_hi"`
	LonHi         float64 `yaml:"lng_hi"`
	MinPriceLimit float64 `yaml:"min_price_limit"`
	MaxPriceLimit float64 `yaml:"max_price_limit"`
	HeatmapStep   float64 `yaml:"heatmap_step"`
	Path          string
}

type pricePoint struct {
	id         int
	livingArea int
	price      int
	priceSqm   float64
	lat        float64
	lon        float64
}

func parseConfig(path string) (dataConfig, error) {
	config_text, err := ioutil.ReadFile(path)
	if err != nil {
		return dataConfig{}, err
	}
	config := dataConfig{}
	err = yaml.Unmarshal([]byte(config_text), &config)
	if err != nil {
		return dataConfig{}, err
	}
	config.Path = filepath.Dir(path)
	return config, nil
}

func loadPrices(dataFile *os.File) ([]pricePoint, error) {
	pricePoints := []pricePoint{}
	r := csv.NewReader(dataFile)

	// Read header and throw it away
	if _, err := r.Read(); err != nil {
		return nil, err
	}

	var id, livingArea, price int
	var lat, lon float64
	for {
		rec, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		id, err = strconv.Atoi(rec[0])
		livingArea, err = strconv.Atoi(rec[1])
		price, err = strconv.Atoi(rec[2])
		lat, err = strconv.ParseFloat(rec[3], 64)
		lon, err = strconv.ParseFloat(rec[4], 64)
		if err != nil {
			return nil, err
		}
		if livingArea != 0 {
			pricePoint := pricePoint{
				id:         id,
				livingArea: livingArea,
				price:      price,
				priceSqm:   float64(price) / float64(livingArea),
				lat:        lat,
				lon:        lon,
			}
			pricePoints = append(pricePoints, pricePoint)
		}
	}
	return pricePoints, nil
}

func main() {
	var err error
	var config dataConfig
	var priceData []pricePoint
	dataFilePath := flag.String("d", "", "Path to input data")
	outputPath := flag.String("o", "/.", "Path to output")
	configFilePath := flag.String("c", "config.yml", "Path to config file")
	printBuckets := flag.Bool("b", false, "Prints bucket list to output")
	flag.Parse()

	if config, err = parseConfig(*configFilePath); err != nil {
		log.Fatalf("Not a valid configuration file, %v", err)
	}

	heatmapConfig := HeatMapConfig{
		Width:    config.HeatmapWidth,
		Height:   config.HeatmapHeight,
		LatHi:    config.LatHi,
		LatLo:    config.LatLo,
		LonHi:    config.LonHi,
		LonLo:    config.LonLo,
		MaxLimit: config.MaxPriceLimit,
		MinLimit: config.MinPriceLimit,
		Step:     config.HeatmapStep,
	}

	heatmap := NewHeatmap(heatmapConfig)

	if *printBuckets {
		// Print color buckets and exit
		buckets := heatmap.GetBucketList()
		b, _ := json.Marshal(buckets)
		ioutil.WriteFile(*outputPath, b, 0644)
		os.Exit(0)
	}

	if *dataFilePath != "" {
		dataFile, err := os.Open(*dataFilePath)
		if err != nil {
			log.Fatalf("Could not open file: %s", *dataFilePath)
		}
		defer dataFile.Close()
		priceData, err = loadPrices(dataFile)
	} else {
		info, _ := os.Stdin.Stat()
		if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
			log.Fatalf("Either specify a data file via the -c flag or pipe it to the application\n")
		}
		priceData, err = loadPrices(os.Stdin)
	}

	log.Printf("%d records has been loaded", len(priceData))

	// Generate the heatmap
	img := heatmap.Create(priceData)
	imgFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Could not create file: %s", *outputPath)
	}
	defer imgFile.Close()
	png.Encode(imgFile, img)
	fmt.Printf("Saved image: %s\n", *outputPath)
}
