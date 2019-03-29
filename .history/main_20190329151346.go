package main

import (
        "fmt"
        "strings"
        "time"

        "github.com/turnage/graw"
        "github.com/turnage/graw/reddit"
)

type spBot struct {
        bot reddit.Bot
}

func (r *spBot) Post(p *reddit.Post) error {
        if strings.Contains(p.SelfText, "remind me of this post") {
                <-time.After(10 * time.Second)
                return r.bot.SendMessage(
                        p.Author,
                        fmt.Sprintf("Reminder: %s", p.Title),
                        "You've been reminded!",
                )
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