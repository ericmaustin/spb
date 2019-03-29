package main

import (
	"strings"
	// "time"
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
	// log.Debugf("Found post:\n%sby%s", p.SelfText, p.Author)

	// if strings.Contains(p.SelfText, "post") {
	// 	log.Debugf("Found post that matches:\n%v", p.SelfText)
	// }
	return nil
}

func (r *spBot) Comment(post *reddit.Comment) error {
	log.Debugf("Found comment:\n%sby%s", post.Body, post.Author)

	if strings.Contains(c.Body, "post") {
		log.Debugf("Found post that matches:\n%v", c.Body)
	}
	return nil
}

func main() {

	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)

	if bot, err := reddit.NewBotFromAgentFile("bot.agent", 0); err != nil {
		log.Errorf("Failed to create bot handle: %v", err)
	} else {
		cfg := graw.Config{Subreddits: []string{"funny"}}
		handler := &spBot{bot: bot}
		if _, wait, err := graw.Run(handler, bot, cfg); err != nil {
			log.Errorf("Failed to start graw run: %v", err)
		} else {
			log.Errorf("Graw run failed with error %v", wait())
		}
	}
}