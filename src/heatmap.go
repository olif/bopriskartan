package main

// heatmap.go takes a slice of pricepoints and converts them to a
// png image where each pixel is calculated as a weighted distance
// mean.
//
// (0,0) is (lng_lo, lat_hi) and (max_x,max_y) is (lng_hi, lat_lo)
//
// (lng_lo,lat_hi) 0 -------------> (lng_hi, lat_hi)
//                 |
//                 |
//                 |
//                 ↓                (lng_hi, lat_lo)
//

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"
)

const (
	ignoreDist = 0.01
	stdDev     = ignoreDist / 2
	workers    = 8
)

var twoStdDevSqr = stdDev * stdDev * 2

type Heatmap struct {
	config           HeatMapConfig
	lonDist          float64
	latDist          float64
	nrOfColorBuckets int
}

type HeatMapConfig struct {
	Width    int
	Height   int
	LatHi    float64
	LonHi    float64
	LatLo    float64
	LonLo    float64
	MaxLimit float64
	MinLimit float64
	Step     float64
}

type ColorBucket struct {
	Price float64
	Color color.NRGBA
}

func NewHeatmap(c HeatMapConfig) *Heatmap {
	// Calculate the number of color buckets. Add two extra for all values below
	// and above the limits
	totalRange := c.MaxLimit - c.MinLimit
	nrOfBuckets := int(totalRange/c.Step) + 2
	return &Heatmap{
		config:           c,
		lonDist:          c.LonHi - c.LonLo,
		latDist:          c.LatHi - c.LatLo,
		nrOfColorBuckets: nrOfBuckets,
	}
}

// Convert a lat/lon pair to corresponding cartesian coordinate value.
func (g *Heatmap) llToPixel(lon float64, lat float64) (int, int, error) {
	fracLon := (lon - g.config.LonLo) / g.lonDist
	fracLat := (g.config.LatHi - lat) / g.latDist

	x := float64(g.config.Width) * fracLon
	y := float64(g.config.Height) * fracLat

	return int(x), int(y), nil
}

// Convert a cartesian coordinate value to lat/lng
func (g *Heatmap) pixelToLL(x int, y int) (float64, float64, error) {

	// x is lng, y is lat
	fracX := float64(x) / float64(g.config.Width)
	fracY := float64(y) / float64(g.config.Height)

	lon := g.config.LonLo + fracX*g.lonDist
	lat := g.config.LatHi - fracY*g.latDist

	return lon, lat, nil
}

func distanceSquared(x1, x2, y1, y2 float64) float64 {
	return (x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)
}

func calcPrice(prices []pricePoint, lon, lat float64) (float64, error) {
	var num, dnm float64
	c := 0
	for _, pricePoint := range prices {
		if pricePoint.priceSqm != 0 {
			distanceSqrd := distanceSquared(lon, pricePoint.lon, lat, pricePoint.lat)
			weight := (1 / (twoStdDevSqr * math.Pi)) * math.Exp(distanceSqrd*(-1/twoStdDevSqr))

			if weight > 2 {
				c += 1
			}

			num += pricePoint.priceSqm * weight
			dnm += weight
		}
	}

	if c <= 2 {
		return -1, nil
	}

	return num / dnm, nil
}

// The color buckets are zero indexed
func (k *Heatmap) getBucketForPrice(price float64) int {
	if price > k.config.MaxLimit {
		return k.nrOfColorBuckets - 1
	} else if price < k.config.MinLimit {
		return 0
	} else {
		return int((price-k.config.MinLimit)/k.config.Step) + 1
	}
}

// priceToColor returns a RGB color for a given price. The color is first
// created in HSL format since we then make a simple linear interpolation
// to calculate the gradient. We interpolate between blue (h = 250) to red
// (h = 0)
func (k *Heatmap) priceToColor(price float64) color.NRGBA {
	bucket := k.getBucketForPrice(price)
	if price == -1 {
		return color.NRGBA{255, 255, 255, 0}
	}
	q := float64(bucket) / float64(k.nrOfColorBuckets)
	h := (290.0 * (1.0 - q)) - 40
	if h < 0 {
		h = 360 + h
	}
	h = h / 360.0
	r, g, b := hslToRgb(h, 1, 0.5)
	return color.NRGBA{r, g, b, 255}
}

func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	var rPrim, gPrim, bPrim float64
	if s == 0 {
		rPrim = l
		gPrim = l
		bPrim = l
	} else {
		var q, p float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p = 2*l - q
		rPrim = hueToRgb(p, q, h+(1.0/3.0))
		gPrim = hueToRgb(p, q, h)
		bPrim = hueToRgb(p, q, h-(1.0/3.0))
	}
	return uint8(rPrim * 255), uint8(gPrim * 255), uint8(bPrim * 255)
}

func hueToRgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6.0
	}
	return p
}

type task struct {
	x     int
	y     int
	price float64
}

// Returns a list of price/color buckets
func (k *Heatmap) GetBucketList() []ColorBucket {
	var bucket ColorBucket
	var buckets = make([]ColorBucket, k.nrOfColorBuckets)

	// Add first bucket for prices less than minlimit
	buckets[0] = ColorBucket{Price: 0, Color: k.priceToColor(0)}

	for i := 1; i < k.nrOfColorBuckets; i++ {
		price := k.config.Step*float64(i-1) + k.config.MinLimit
		bucket = ColorBucket{Price: price, Color: k.priceToColor(price)}
		buckets[i] = bucket
	}
	return buckets
}

// Create a heatmap given a slice of price points
func (h *Heatmap) Create(p []pricePoint) *image.NRGBA {
	var wg sync.WaitGroup
	matrix := make([][]float64, h.config.Height)
	in := make(chan task)
	out := make(chan task)

	wg.Add(h.config.Width * h.config.Width)
	// Create workers
	for w := 0; w < workers; w++ {
		go func() {
			for t := range in {
				lon, lat, _ := h.pixelToLL(t.x, t.y)
				price, _ := calcPrice(p, lon, lat)
				out <- task{t.x, t.y, price}
			}
		}()
	}

	// Initialize a collector
	go func() {
		counter := 0.0
		progress := 0
		lastProgress := 0
		for i := range out {
			if matrix[i.y] == nil {
				matrix[i.y] = make([]float64, h.config.Width)
			}
			matrix[i.y][i.x] = i.price
			counter += 1
			progress = int((counter / (float64(h.config.Width) * float64(h.config.Height))) * 100.0)
			if progress != lastProgress {
				lastProgress = progress
				fmt.Printf("\rProgress: %d%%", progress)
			}
			wg.Done()
		}
	}()

	// Add work
	for i := 0; i < h.config.Height; i++ {
		for j := 0; j < h.config.Width; j++ {
			in <- task{j, i, 0}
		}
	}

	wg.Wait()
	close(in)
	close(out)
	fmt.Println("")
	img := image.NewNRGBA(image.Rect(0, 0, h.config.Width, h.config.Height))
	for i := 0; i < h.config.Height; i++ {
		for j := 0; j < h.config.Width; j++ {
			c := h.priceToColor(matrix[j][i])
			img.Set(i, j, c)
		}
	}

	// Add pricepoints as 1px dots in the image
	for _, point := range p {
		x, y, _ := h.llToPixel(point.lon, point.lat)
		img.Set(x, y, color.NRGBA{0, 0, 0, 255})
	}

	return img
}
