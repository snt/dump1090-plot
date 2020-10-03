# dump1090-plot

Plot traces of aircraft and radar coverage on Google Maps from CSV that comes from port 30003 of dump1090.

Based on a nice code written in Python [plot1090](https://github.com/Wilm0r/plot1090). 

# build

```sh
go build
```

# Collect CSVs

Collect CVSs like this(assuming `dump1090` is running on localhost.):

```sh
while true; do timeout 24h nc localhost 30003 > adsb-24h-$(date +%Y-%m-%d_%H.%M).csv; done
```

# run

Run the code:

```sh
./dump1090-plot -apikey <your-api-key> -lat 35.123456 -lon 139.123456 adsb-24h-input-2020-10-02_07.36.csv
```

* `-apikey` is Google Maps API key.
* `-lat` and `-lon` is the latitude, and the longitude where your dump1090 receiver is.

