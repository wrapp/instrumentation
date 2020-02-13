package lastseen

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
)

type redisExporter struct {
	client     *redis.Client
	sender     string
	defaultKey string
}

// Redis creates a new last seen exporter and connects to the provided redis host
func Redis(host string, sender string, defaultKey string) (Exporter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0, // use default DB
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	exporter := redisExporter{
		client:     client,
		defaultKey: defaultKey,
		sender:     sender,
	}
	return &exporter, nil
}

func (t *redisExporter) Flush(ctx context.Context, field string, val LastSeen) error {
	_, err := t.client.HSet(t.defaultKey, fmt.Sprintf("%s/%s", t.sender, field), val.Value).Result()
	return err
}
