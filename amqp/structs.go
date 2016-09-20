package amqp

type StartedMessage struct {
	GID  string `json:"gid"`
	Time string `json:"time"`
}
