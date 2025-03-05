package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Output struct {
	Timestamp    time.Time `json:"timestamp"`
	LogName      string    `json:"logname"`
	RandomMetric float32   `json:"some_metric"`
}

var (
	tickerTime = 30 * time.Second
)

func main() {
	tick := time.NewTicker(tickerTime)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	loop := true
	for loop {
		select {
		case <-tick.C:
			out := Output{
				Timestamp:    time.Now(),
				LogName:      "json_outputer",
				RandomMetric: r.Float32(),
			}
			b, err := json.Marshal(out)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("%s\n", string(b))
			}
		case <-signalCh:
			loop = false
		}
	}
}
