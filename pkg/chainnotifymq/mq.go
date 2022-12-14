package chainnotifymq

import (
	"context"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/pkg/chainnotifymq/mqmessage"
)

var mqMap = make(map[string]MQ)

func register(mqs ...MQ) {
	for _, mq := range mqs {
		mqMap[mq.Name()] = mq
	}
}

// GetMQ get mq by name
func GetMQ(name string) MQ {
	return mqMap[name]
}

type MQ interface {
	Name() string
	InitFromConfig(ctx context.Context, cfg *config.ChainNotify, queueName string) error
	FetchMessage(ctx context.Context) (mqmessage.Message, error)
	Close() error
}

type CommittableMQ interface {
	MQ
	Commit(ctx context.Context, message mqmessage.Message) error
}
