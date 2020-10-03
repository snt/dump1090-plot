package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kellydunn/golang-geo"
	"io"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"text/template"
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

func main() {
	// TODO use other command line parse library to make some flags mandatory.
	lat := flag.Float64("lat", 0, "latitude of your receiver")
	lon := flag.Float64("lon", 0, "longitude of your receiver")
	destDirectory := flag.String("dest-dir", ".", "HTML output directory.")
	apiKey := flag.String("apikey", "", "API Key of Google Maps")
	flag.Parse()
	inputBstCsvFiles := flag.Args()

	center := Pos{Lat: *lat, Lng: *lon}

	plotFiles(*destDirectory, inputBstCsvFiles, center, *apiKey)
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
	csvf, err := os.Open(csvFn)
	if err != nil {
		log.Printf("failed to open input CSV: %s\n", csvFn)
		return err
	}
	defer csvf.Close()

	htmlf, err := os.Create(htmlFn)
	if err != nil {
		log.Printf("failed to open HTML for write: %s\n", htmlFn)
		return err
	}
	defer htmlf.Close()

	reader := csv.NewReader(csvf)

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

		//alt, err := strconv.ParseInt(cs[11], 10, 32)
		//if err != nil {
		//	continue
		//}

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
	}));`, csvFn, apiKey, htmlf)

	return nil
}

func updateTrack(aircraftMap map[int32]AircraftTrack, tracks [][]Pos, id int32, lat float64, lon float64) [][]Pos {
	pt := geo.NewPoint(lat, lon)
	if ac, ok := aircraftMap[id]; ok {
		last := ac.Position[len(ac.Position)-1]
		distanceKm := pt.GreatCircleDistance(&last)
		az := last.BearingTo(pt)

		if distanceKm > 2.5 {
			//split track if it is 2.5+ km away from previous position
			ps := make([]Pos, 0, len(ac.Position))
			for _, p := range ac.Position {
				ps = append(ps, Pos{Lat: p.Lat(), Lng: p.Lng()})
			}

			tracks = append(tracks, ps)

			ac.Position = []geo.Point{}
			ac.HasAzimuth = false

		} else if len(ac.Position) >= 2 && ac.HasAzimuth {
			//remove previous point to reduce plot data if the a/c is almost on the straight line
			if math.Mod(math.Abs(ac.Azimuth-az), 360.0) < 2.5 {
				ac.Position = ac.Position[:len(ac.Position)-1]
			}

		}

		ac.Position = append(ac.Position, *pt)
		ac.Azimuth = az
		ac.HasAzimuth = true

		aircraftMap[id] = ac
	} else {
		aircraftMap[id] = AircraftTrack{
			Position:   []geo.Point{*pt},
			Azimuth:    0,
			HasAzimuth: false,
		}
	}

	return tracks
}

func updateEdges(dists []Dist, lat float64, lon float64, center *geo.Point) {
	pt := geo.NewPoint(lat, lon)
	distance := center.GreatCircleDistance(pt)
	az := center.BearingTo(pt)

	idx := int(math.Round(float64(len(dists))*math.Mod(az+360.0, 360.0)/360)) % len(dists)
	if distance > dists[idx].Km {
		dists[idx] = Dist{
			Km:  distance,
			Pos: Pos{Lat: lat, Lng: lon},
		}
	}
}

func writeHtml(vars *map[string]interface{}, js string, title string, apiKey string, writer io.Writer) {
	t, err := template.New("html").Funcs(template.FuncMap{
		"json": func(x interface{}) string {
			jb, err := json.Marshal(x)
			if err != nil {
				log.Fatal(err)
			}
			return string(jb)
		},
	}).Parse(outHtmlTemplate)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(writer, struct {
		Title  string
		Vars   map[string]interface{}
		Js     string
		ApiKey string
	}{
		Title:  title,
		Vars:   *vars,
		Js:     js,
		ApiKey: apiKey,
	})
	if err != nil {
		log.Fatal(err)
	}

}

var outHtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="initial-scale=1.0, user-scalable=no">
	<meta charset="utf-8">
	<title>{{.Title}}</title>
	<style>
		html, body {
			height: 100%;
			margin: 0;
			padding: 0;
		}
		#map {
			height: 100%;
		}
	</style>
</head>
<body>
	<div id="map"></div>
	<script>
function initMap() {
	{{range $k, $v := .Vars}}
	const {{$k}} = {{json $v}};
	{{end}}

	// Create the map.
	const map = new google.maps.Map(document.getElementById('map'), {
		zoom: 8,
		center: d.center,
		mapTypeId: google.maps.MapTypeId.TERRAIN
	});
{{.Js}}
}
	</script>
	<script async defer
	    src="https://maps.googleapis.com/maps/api/js?key={{.ApiKey}}&callback=initMap"></script>
	</body>
</html>
`
