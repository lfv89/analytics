package configs

type Config struct {
	Elastic struct {
		URL string `default:"http://elastic:9200" envconfig:"ELASTIC_URL"`
	}

	Web struct {
		NotifyURL string `default:"http://localhost:4002/notify" envconfig:"WEB_NOTIFY_URL"`
	}
}
