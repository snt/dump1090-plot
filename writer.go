package main

import (
	"encoding/json"
	"io"
	"log"
	"text/template"
)

type HtmlArgs struct {
	Title  string
	Vars   map[string]interface{}
	Js     string
	ApiKey string
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

	err = t.Execute(writer, HtmlArgs{
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
