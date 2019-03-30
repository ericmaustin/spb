package main

import (
	// "strings"
	// "time"
	"fmt"
	"os"
	"regexp"

	logging "github.com/op/go-logging"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

type spBot struct {
	bot reddit.Bot
}

var (
	log = logging.MustGetLogger("spBot")

	Text = `**If you're in crises, please consider contacting the [National Suicide Prevention Lifeline](https://suicidepreventionlifeline.org/): 1-800-273-TALK**

If you wish to help others in crises consider supporting:
- [The National Suicide Prevention Lifeline](https://suicidepreventionlifeline.org/donate/)
- [The National Alliance on Mental Illness](https://donate.nami.org/give/197406/#!/donation/checkout)
- [The Jason Foundation](http://jasonfoundation.com/get-involved/)

	
I am a bot created by a survivor that is in no way affiliated with any of the organizations mentioned | [feedback](mailto:erics.awesome.bots@gmail.com)
`

	matchExpressions = []string{
		`i\s*(am\sgoing\sto|will)\skill\smyself`,
		`(will|am\sgoing\sto|want\sto)\scommit\ssuicide`,
		`thinking.*about.*(suicide|killing\smyself)`,
		`contemplating\s+suicide`,
	}

	matchRe = make(map[string]*regexp.Regexp, len(matchExpressions))

	blacklist = []string{
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
)

// isBlackCommentBlackListed checks if comment has anything that reports it as blacklisted
func (r *spBot) isBlackCommentBlackListed(post *reddit.Comment) bool {
	for _, s := range blacklist {
		if post.Subreddit == s || post.SubredditID == s {
			log.Debugf("Ignoring matching comment in subreddit: '%s':\nbody: %s\nby: %s",
				s, post.Body, post.Author)
			return true
		}
	}

	return false
}

func (r *spBot) Comment(post *reddit.Comment) error {

	for reString, re := range matchRe {

		if re.MatchString(post.Body) {
			if r.isBlackCommentBlackListed(post) {
				return nil
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
	backend1Formatter := logging.NewBackendFormatter(backend1,
		logging.MustStringFormatter(`%{color}%{time:Jan _2 15:04:05.000} %{level:.4s} ▶%{color:reset} %{message}`))

	fh, err := os.OpenFile("bot.log", os.O_CREATE | os.O_APPEND | os.O_RDWR, 0666)

	if err != nil {
		fmt.Printf("Error opening log file: %v", err)
		return
	}

	backendFile := logging.NewLogBackend(fh, "", 0)
	backendFileFormatter := logging.NewBackendFormatter(backendFile,
		logging.MustStringFormatter(`%{time:Jan _2 15:04:05.000} %{level:.4s} ▶ %{message}`))

	logging.SetBackend(backend1Formatter, backendFileFormatter)

	log.Infof("Started suicide prevention bot.")

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
