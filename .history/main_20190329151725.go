package main

import (
        "fmt"
        "strings"
        "time"

		"github.com/op/go-logging"
        "github.com/turnage/graw"
        "github.com/turnage/graw/reddit"
)

type spBot struct {
        bot reddit.Bot
}

var log = logging.MustGetLogger("spBot")

func (r *spBot) Post(p *reddit.Post) error {
        if strings.Contains(p.SelfText, "post") {
                
        }
        return nil
}

func main() {
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