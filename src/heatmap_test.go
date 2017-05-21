package main

import (
	"testing"
)

var testConfig = HeatMapConfig{
	Width:    1000,
	Height:   1000,
	LatHi:    59.439306,
	LatLo:    59.205634,
	LonHi:    18.266219,
	LonLo:    17.846684,
	MaxLimit: 50000,
	MinLimit: 10000,
	Step:     10000,
}

var heatmap = NewHeatmap(testConfig)

type testTableEntry struct {
	x   int
	y   int
	lon float64
	lat float64
}

func TestPixelToLL(t *testing.T) {
	testTable := []testTableEntry{
		{0, 0, testConfig.LonLo, testConfig.LatHi},
		{testConfig.Width, testConfig.Height, testConfig.LonHi, testConfig.LatLo},
	}

	for _, entry := range testTable {
		lon, lat, _ := heatmap.pixelToLL(entry.x, entry.y)
		if lon != entry.lon || lat != entry.lat {
			t.Errorf("Conversion out of range. Lon: %g (%g), Lat: %g (%g)", lon, entry.lon, lat, entry.lat)
		}
	}
}

func TestLLToPixel(t *testing.T) {
	testTable := []testTableEntry{
		{0, 0, testConfig.LonLo, testConfig.LatHi},
		{testConfig.Width, testConfig.Height, testConfig.LonHi, testConfig.LatLo},
	}

	for _, entry := range testTable {
		x, y, _ := heatmap.llToPixel(entry.lon, entry.lat)
		if x != entry.x || y != entry.y {
			t.Errorf("Conversion out of range. X: %d (%d), Y: %d (%d)", x, entry.x, y, entry.y)
		}
	}
}

func TestHslToRgb(t *testing.T) {
	r, g, b := hslToRgb((359 / 360), 1, 1)
	if r != 255 || g != 255 || b != 255 {
		t.Errorf("HSL to rgb failed: R: %d (%d), G: %d (%d), B: %d (%d)", 255, r, 255, g, 255, b)
	}

	r, g, b = hslToRgb((200.0 / 360.0), 1, 0.5)
	if r != 0 || g != 169 || b != 255 {
		t.Errorf("HSL to rgb failed: R: %d (%d), G: %d (%d), B: %d (%d)", 0, r, 169, g, 255, b)
	}
}

func TestGetBucketList(t *testing.T) {
	heatmap := NewHeatmap(testConfig)
	buckets := heatmap.GetBucketList()
	validKeys := []float64{0, 10000, 20000, 30000, 40000, 50000}
	for i, _ := range validKeys {
		validBucketValue := validKeys[i]
		actualBucketValue := buckets[i]
		if validBucketValue != actualBucketValue.Price {
			t.Errorf("Expected key: %f not found in bucket list", validBucketValue)
		}
	}
}

func TestBucketForPrice(t *testing.T) {
	heatmap := NewHeatmap(testConfig)
	testTable := map[float64]int{
		5000:  0,
		10000: 1,
		20000: 2,
		30000: 3,
		40000: 4,
		50000: 5,
		55000: 5}

	for k, v := range testTable {
		bucket := heatmap.getBucketForPrice(k)
		if bucket != v {
			t.Errorf("Expected bucket: %d for price: %f, but got: %d", v, k, bucket)
		}
	}
}
