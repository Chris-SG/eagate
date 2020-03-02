module github.com/chris-sg/eagate

go 1.13

require (
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/chris-sg/eagate_models v0.0.0
	github.com/jmoiron/sqlx v1.2.1-0.20191203222853-2ba0fc60eb4a // indirect
	github.com/lib/pq v1.3.0 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
)

replace github.com/chris-sg/eagate_models v0.0.0 => ../eagate_models
