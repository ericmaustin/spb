package main

import (
	"strings"
	"time"
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

	
^(I am a bot created by a survivor that is in no way affiliated with any of the organizations mentioned |) [^(feedback)](mailto:erics.awesome.bots@gmail.com)
`

	matchExpressions = []string{
		`(?i)i\s*(am\sgoing\sto|will|will\sbe)\skill(ing)?\smyself`,
		`(?i)(will|am\sgoing\sto|want\sto)\s(commit\ssuicide|kill\smyself)`,
		`(?i)thinking.*about.*(suicide|killing\smyself)`,
		`(?i)(contemplating|considering|thinking\sabout)\s+suicide`,
		`(?i)(planning\sto\s|have\splans\sto)\s+(commit\ssuicide|kill\smyself)`,
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

// isCommentBlackListed checks if comment has anything that reports it as blacklisted
func (r *spBot) isCommentBlackListed(post *reddit.Comment) bool {
	for _, s := range blacklist {
		if post.Subreddit == s || post.SubredditID == s {
			log.Debugf("Ignoring matching comment:\nsubreddit: %s\nBody: %s\nby: %s",
				post.Subreddit, post.Body, post.Author)
			return true
		}
	}

	return false
}

// isPostBlackListed checks if post has anything that reports it as blacklisted
func (r *spBot) isPostBlackListed(post *reddit.Post) bool {
	for _, s := range blacklist {
		if post.Subreddit == s || post.SubredditID == s {
			log.Debugf("Ignoring matching post:\nsubreddit: %s\ntitle: %s\nSelf Text: %s\nby: %s",
				post.Subreddit, post.Title, post.SelfText, post.Author)
			return true
		}
	}

	return false
}

func (r *spBot) Comment(post *reddit.Comment) error {

	for reString, re := range matchRe {
		if re.MatchString(post.Body) {
			if r.isCommentBlackListed(post) {
				return nil
			}
			log.Debugf("Found matching post for expression '%s':\nsubreddit: %s\nBody: %s\nby: %s",
				reString, post.Subreddit, post.Body, post.Author)
			r.bot.Reply(post.Name, Text)
			continue
		}
	}

	return nil
}

func (r *spBot) Post(post *reddit.Post) error {

	for reString, re := range matchRe {

		if re.MatchString(post.Title) || re.MatchString(post.SelfText) {
			if r.isPostBlackListed(post) {
				return nil
			}

			log.Debugf("Found matching post for expression '%s':\nsubreddit: %s\ntitle: %s\nSelf Text: %s\nby: %s",
				reString, post.Subreddit, post.Title, post.SelfText, post.Author)
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
		cfg := graw.Config{SubredditComments: []string{"all"}, Subreddits:[]string{"all"}}
		handler := &spBot{bot: bot}
		if _, wait, err := graw.Run(handler, bot, cfg); err != nil {
			log.Errorf("Failed to start graw run: %v", err)
		} else {
			errStr := fmt.Sprintf("%v", wait())

			if strings.Contains(errStr, "bad response code: 500") {
				log.Warningf("Restarting bot. Graw got a 500 error: %s", errStr)
				time.Sleep(time.Millisecond * 500)
				main()
			}

			if strings.Contains(errStr, "token expired") {
				log.Warningf("Restarting bot. Graw got a token error: %s", errStr)
				main()
			}

			log.Errorf("Graw run failed with error %s", errStr)
		}
	}
}
