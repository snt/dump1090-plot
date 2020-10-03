package main

import (
	geo "github.com/kellydunn/golang-geo"
	"math"
)

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

