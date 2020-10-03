package main

import (
	"encoding/csv"
	"fmt"
	geo "github.com/kellydunn/golang-geo"
	"io"
	"log"
	"os"
	"path"
	"strconv"
)

type AircraftTrack struct {
	Position   []geo.Point
	Azimuth    float64
	HasAzimuth bool
}

type Pos struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Dist struct {
	Km  float64
	Pos Pos
}

func plotFiles(destDirectory string, inputBstCsvFiles []string, center Pos, apiKey string) {
	for _, fn := range inputBstCsvFiles {
		outFn := path.Join(destDirectory, fmt.Sprintf("plot-%s.html", path.Base(fn)))
		err := plot(fn, outFn, center, apiKey)
		if err != nil {
			log.Printf("skipping csv %s\n", fn)
		}
	}
}

func plot(csvFn string, htmlFn string, center Pos, apiKey string) error {
	csvFile, err := os.Open(csvFn)
	if err != nil {
		log.Printf("failed to open input CSV: %s\n", csvFn)
		return err
	}
	defer csvFile.Close()

	htmlFile, err := os.Create(htmlFn)
	if err != nil {
		log.Printf("failed to open HTML for write: %s\n", htmlFn)
		return err
	}
	defer htmlFile.Close()

	reader := csv.NewReader(csvFile)

	aircraftMap := make(map[int32]AircraftTrack)

	var tracks [][]Pos

	var dists = make([]Dist, 360*4)

	for {
		cs, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("failed to parse csv")
			return err
		}

		// we need messages with include lat and lon (MSG,2, ... and MSG,3, ...)
		if !(cs[0] == "MSG" && (cs[1] == "2" || cs[1] == "3")) {
			continue
		}

		id64, err := strconv.ParseInt(cs[4], 16, 32)
		if err != nil {
			continue
		}
		id := int32(id64)

		lat, err := strconv.ParseFloat(cs[14], 64)
		if err != nil {
			continue
		}

		lon, err := strconv.ParseFloat(cs[15], 64)
		if err != nil {
			continue
		}

		tracks = updateTrack(aircraftMap, tracks, id, lat, lon)
		gpc := geo.NewPoint(center.Lat, center.Lng)
		updateEdges(dists, lat, lon, gpc)
	}

	vars := make(map[string]interface{})
	vars["polys"] = tracks
	vars["d"] = map[string]interface{}{
		"center": center,
	}

	edges := make([]Pos, len(dists))
	for i, d := range dists {
		if d.Km == 0 {
			edges[i] = center
		} else {
			edges[i] = d.Pos
		}
	}

	vars["edges"] = edges

	writeHtml(&vars, `
	const rangeCircles = [100,150,200]
		.map(r => r * 1852)
		.map(r => new google.maps.Circle({
			map,
			center: d.center,
			radius: r,
			strokeColor: '#000000',
			strokeOpacity: 0.5,
			strokeWeight: 1,
			fillColor: '#000000',
			fillOpacity: 0
		}));

	const edgePolygon = new google.maps.Polygon({
		path: edges,
		strokeColor: '#ffff00',
		strokeOpacity: 0.8,
		strokeWeight: 1,
		fillColor: '#ffff00',
		fillOpacity: 0.3,
		map: map
	});

	const trackPolyLines = polys.map(p => new google.maps.Polyline({
		path: p,
		strokeColor: '#ff0000',
		strokeOpacity: 0.25,
		strokeWeight: 1,
		map
	}));`, csvFn, apiKey, htmlFile)

	return nil
}
