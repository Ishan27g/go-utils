package example

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

var opts []nats.Option
var urls = "nats://localhost:4222"

var subjects map[string]*string
var lock = sync.RWMutex{}

type SubCb func()

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 5 * time.Second
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Disconnected due to:%s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}
func init() {

	subjects = make(map[string]*string)
	lock = sync.RWMutex{}
	// Connect Options.
	log.SetFlags(log.LstdFlags)
	opts := []nats.Option{nats.Name("-nats-")}
	opts = setupConnOptions(opts)

}

func sub(subj string, cb func(ctx context.Context, msg *nats.Msg) bool) {
	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		log.Println(err)
		return
	}
	nc.Subscribe(subj, func(msg *nats.Msg) {
		cb(context.Background(), msg)
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on [%s]", subj)

}

func NatsSubscribe(subj string, cb func(ctx context.Context, msg *nats.Msg) bool) {
	lock.Lock()
	if subjects[subj] == nil {
		subjects[subj] = &subj
	}
	lock.Unlock()
	sub(subj, cb)
}
func NatsPublish(subj string, msg string, reply *string) bool {

	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		log.Println(err)
		return false
	}
	defer nc.Close()
	lock.Lock()
	if subjects[subj] == nil {
		subjects[subj] = &subj
	}
	lock.Unlock()
	if reply != nil && *reply != "" {
		err = nc.PublishRequest(subj, *reply, []byte(msg))
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	} else {
		err = nc.Publish(subj, []byte(msg))
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}

	err = nc.Flush()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Published [%s] : '%s'\n", subj, msg)
	}
	return true
}
