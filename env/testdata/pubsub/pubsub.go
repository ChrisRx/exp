package pubsub

type Config struct {
	Topic        string `env:"PUBSUB_TOPIC"`
	Subscription string `env:"PUBSUB_SUBSCRIPTION"`
}
