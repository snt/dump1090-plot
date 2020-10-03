package main

import (
	geo "github.com/kellydunn/golang-geo"
	"math"
)

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
