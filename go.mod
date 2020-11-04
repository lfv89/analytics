module github.com/lfv89/analytics

go 1.14

replace github.com/lfv89/analytics v0.0.0-20201118004215-8ee4b70861bd => ../analytics

require (
	github.com/elastic/go-elasticsearch/v7 v7.10.0
	github.com/gorilla/websocket v1.4.2
	github.com/kelseyhightower/envconfig v1.4.0
)
