package private

type Event struct {
	Source    string `json:"source"`
	ClientID  int    `json:"clientID"`
	UserAgent string `json:"userAgent"`
}
