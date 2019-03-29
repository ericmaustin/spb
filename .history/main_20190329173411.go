package main

import (
	// "strings"
	// "time"
	"os"
	"regexp"

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

var Text = `If you're in crises, please consider contacting the **[National Suicide Prevention Lifeline](https://suicidepreventionlifeline.org/): 1-800-273-TALK***

If you wish to help others in crises consider supporting:
- [The National Suicide Prevention Lifeline](https://suicidepreventionlifeline.org/donate/)
- [The National Alliance on Mental Illness](https://donate.nami.org/give/197406/#!/donation/checkout)
- [The Jason Foundation](http://jasonfoundation.com)

	
^I am a bot created by a survivor that is in no way affliated with any of the organizations mentioned. Please send feedback to [this addres](mailto:ericmaustin+spb@gmail.com)
`

var matchExpressions = []string{
	`i\s*(am\sgoing\sto|will)\skill\smyself`,
	`(will|am\sgoing\sto|want\sto)\scommit\ssuicide`,
	`thinking.*about.*(suicide|killing\smyself)`,
	`contemplating\s+suicide`,
}

var matchRe = make(map[string]*regexp.Regexp, 0)

var blacklist = []string{
	// music subrreddits with potential false positives for lyrics
	`music`,
	`listentothis`,
	`lyrics`,
	`metal`,
	`hiphopheads`,
	`indieheads`,
	`edm`,
	`mixes`,
	// quote subreddits
	`quotes`,
}

// func (r *spBot) Post(p *reddit.Post) error {
// 	log.Debugf("New post:%sby%s", p.Title, p.Author)

// 	// if strings.Contains(p.SelfText, "post") {
// 	// 	log.Debugf("Found post that matches:\n%v", p.SelfText)
// 	// }
// 	return nil
// }

func (r *spBot) Comment(post *reddit.Comment) error {

	for reString, re := range matchRe {

		if re.MatchString(post.Body) {
			for _, s := range blacklist {
				if post.Subreddit == s || post.SubredditID == s {
					log.Debugf("Ignoring matching comment for expression '%s' in subreddit: '%s':\nbody: %s\nby: %s", 
						s, reString, post.Body, post.Author)
					return nil
				}
			}

			log.Debugf("Found matching comment for expression '%s':\nbody: %s\nby: %s", 
				reString, post.Body, post.Author)
			r.bot.Reply(post.Name, Text)
		}
	}

	return nil
}

func main() {

	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)

	for _, exp := range matchExpressions {
		re, err := regexp.Compile(exp)
		if err != nil {
			log.Errorf("Error compiling regex expression %s", re)
			return
		}
		matchRe[exp] = re
	}

	if bot, err := reddit.NewBotFromAgentFile("bot.agent", 0); err != nil {
		log.Errorf("Failed to create bot handle: %v", err)
	} else {
		cfg := graw.Config{SubredditComments: []string{"ericbottesting"}}
		handler := &spBot{bot: bot}
		if _, wait, err := graw.Run(handler, bot, cfg); err != nil {
			log.Errorf("Failed to start graw run: %v", err)
		} else {
			log.Errorf("Graw run failed with error %v", wait())
		}
	}
}