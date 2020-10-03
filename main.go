package main

import (
	"flag"
)

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
