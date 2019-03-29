package main

import (
        "fmt"
        "strings"
		"time"
		"os"

		"github.com/op/go-logging"
        "github.com/turnage/graw"
        "github.com/turnage/graw/reddit"
)

type spBot struct {
        bot reddit.Bot
}

var log = logging.MustGetLogger("spBot")

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)


func (r *spBot) Post(p *reddit.Post) error {
        if strings.Contains(p.SelfText, "post") {
			
        }
        return nil
}

func main() {

	backend1 := logging.NewLogBackend(os.Stdout, "", 0)

	logging.SetBackend(backend1, logging.NewBackendFormatter(backend2, format))

	if bot, err := reddit.NewBotFromAgentFile("bot.agent", 0); err != nil {
			fmt.Println("Failed to create bot handle: ", err)
	} else {
			cfg := graw.Config{Subreddits: []string{"bottesting"}}
			handler := &spBot{bot: bot}
			if _, wait, err := graw.Run(handler, bot, cfg); err != nil {
					fmt.Println("Failed to start graw run: ", err)
			} else {
					fmt.Println("graw run failed: ", wait())
			}
	}
}