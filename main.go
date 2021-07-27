package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"go.uber.org/atomic"
)

func main() {
	n := flag.Int("n", 1, "count of instances")
	addr := flag.String("addr", "localhost:8090", "metrics addr")
	flag.Parse()

	fmt.Printf("go tool pprof -inuse_space td-overhead http://%s/debug/pprof/heap\n", *addr)

	running := atomic.NewInt32(0)

	staging := dcs.Staging()

	for i := 0; i < *n; i++ {
		go func() {
			defer running.Dec()
			client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
				DCList: staging,
				UpdateHandler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
					return nil
				}),
			})
			log.Fatal(client.Run(context.Background(), func(ctx context.Context) error {
				if err := client.Ping(ctx); err != nil {
					return err
				}
				running.Inc()
				<-ctx.Done()
				return nil
			}))
		}()
	}

	go func() {
		log.Fatal(http.ListenAndServe(*addr, nil))
	}()

	last := int32(0)
	for range time.NewTicker(time.Second * 1).C {
		if v := running.Load(); last != v {
			last = v
			fmt.Printf("running: %d\n", running.Load())
		}
	}
}
