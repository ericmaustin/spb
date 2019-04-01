package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

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

You may also find support by posting in r/SuicideWatch.

If you wish to help others in crises consider supporting:
- [The National Suicide Prevention Lifeline](https://suicidepreventionlifeline.org/donate/)
- [The National Alliance on Mental Illness](https://donate.nami.org/give/197406/#!/donation/checkout)
- [The Jason Foundation](http://jasonfoundation.com/get-involved/)

	
^(I am a bot created by a survivor that is in no way affiliated with any of the organizations mentioned |) [^(feedback)](mailto:erics.awesome.bots@gmail.com)
`

	// matching regex expressions
	matchRe = []*regexp.Regexp{
		regexp.MustCompile(`(?i)i\s*(am\sgoing\sto|will|will\sbe|should|(should\sjust))\skill(ing)?\smyself`),
		regexp.MustCompile(`(?i)(will|am\sgoing\sto|want\sto)\s(commit\ssuicide|kill\smyself)`),
		regexp.MustCompile(`(?i)thinking.*about.*(suicide|killing\smyself)`),
		regexp.MustCompile(`(?i)(contemplating|considering|thinking\sabout)\s+(suicide|killing\smyself)`),
		regexp.MustCompile(`(?i)(planning\sto\s|have\splans\sto)\s+(commit\ssuicide|kill\smyself)`),
	}

	// negating regex expressions
	negateRe = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(i\sam\snot)|(i\sdon'?t\swant\sto)|never|(will\snot)|can'?t|won'?t`),
	}

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
		// suicide watch, not needed there
		`SuicideWatch`,
		// gaming subreddits (lots of false positives)
		`gaming`,
		`GamerNews`,
		`leagueoflegends`,
		`DotA2`,
		`Minecraft`,
		`Skyrim`,
		`Starcraft`,
		`Fallout`,
		`KerbalSpaceProgram`,
		`GrandTheftAutoV`,
		`wow`,
		`MagicTCG`,
		`tf2`,
		`AnimalCrossing`,
		`GameDeals`,
		`ffxiv`,
		`civ`,
		`zelda`,
		`GlobalOffensive`,
		`paydaytheheist`,
	}

	itemExpires = time.Hour * 24
)

// test regex
func testRegex(input string) bool {
	for _, re := range matchRe {
		match_indexes := re.FindStringIndex(input)
		if len(match_indexes) > 0 {
			// make sure nothing NEGATES the match
			prevStr := input[int(math.Max(float64(match_indexes[0]-16), 0)):match_indexes[0]]
			for _, nre := range negateRe {
				if nre.MatchString(prevStr) {
					log.Debugf("testRegex matched %s, but was negated due to this text: %s", input, prevStr)
					return false
				}
			}
			return true
		}
	}
	return false
}

type commentTheadCacheItem struct {
	comment *reddit.Comment
	post    *reddit.Post
	expires time.Time
}

// check if item in the cache has expired
func (i *commentTheadCacheItem) expired() bool {
	return i.expires.Before(time.Now())
}

// cache struct that contains links that have been cached
type threadCache struct {
	items  map[string]*commentTheadCacheItem
	stopCh chan bool
}

// adds an item into the cache
func (tc *threadCache) addComment(comment *reddit.Comment, expiresIn time.Duration) {
	tc.items[comment.LinkURL] = &commentTheadCacheItem{
		comment: comment,
		expires: time.Now().Add(expiresIn),
	}
}

// adds an item into the cache with a post
func (tc *threadCache) addPost(post *reddit.Post, expiresIn time.Duration) {
	tc.items[post.URL] = &commentTheadCacheItem{
		post:    post,
		expires: time.Now().Add(expiresIn),
	}
}

// janitor removes items from the cache as they expire
func (tc *threadCache) janitor() {
	log.Debugf("Cache janitor started.")

	for {
		// poll every 100 milliseconds
		timer := time.NewTicker(time.Millisecond * 100)

		select {
		case stop := <-tc.stopCh:
			if stop == true {
				log.Debugf("Janitor stop request recieved. Exiting.")
				timer.Stop()
				return
			}
		case <-timer.C:
			// we got a tick, purge expired
			for name, item := range tc.items {
				if item.expired() {
					log.Debugf("Item in cache has epxired with URL: %s", name)
					delete(tc.items, name)
				}
			}
		}
	}
}

// stopJanitor stops the janitor
func (tc *threadCache) stopJanitor() {
	tc.stopCh <- true
}

// the actual cache instance
var cache = &threadCache{
	items:  make(map[string]*commentTheadCacheItem, 0),
	stopCh: make(chan bool),
}

// make sure this is the only reply in the thread
func (r *spBot) checkCommentExistsInCache(comment *reddit.Comment) bool {

	if _, ok := cache.items[comment.LinkURL]; ok {
		return true
	}

	return false
}

// make sure this is the only reply in the thread
func (r *spBot) checkPostExistsInCache(post *reddit.Post) bool {

	if _, ok := cache.items[post.URL]; ok {
		return true
	}

	return false
}

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

// Listen to comments
func (r *spBot) Comment(post *reddit.Comment) error {

	if testRegex(post.Body) {
		if r.isCommentBlackListed(post) {
			return nil
		}

		if !r.checkCommentExistsInCache(post) {
			cache.addComment(post, itemExpires)
			log.Infof("Added item to cache, LinkURL: %s", post.LinkURL)
		} else {
			log.Debugf("Found matching comment but LinkURL %s was found in cache :\nsubreddit: %s\nBody: %s\nby: %s",
				post.LinkURL, post.Subreddit, post.Body, post.Author)
			return nil
		}

		log.Infof("Found matching comment:\nLink: %s\nsubreddit: %s\nBody: %s\nby: %s",
			post.Permalink, post.Subreddit, post.Body, post.Author)

		err := r.bot.Reply(post.Name, Text)
		if err != nil {
			log.Errorf("Got error trying to post reply: %v", err)
		}
	}

	return nil
}

// listen to post
func (r *spBot) Post(post *reddit.Post) error {

	if testRegex(post.Title) || testRegex(post.SelfText) {
		if r.isPostBlackListed(post) {
			return nil
		}

		if !r.checkPostExistsInCache(post) {
			cache.addPost(post, itemExpires)
			log.Infof("Added item to cache, LinkURL: %s", post.URL)
		} else {
			log.Debugf("Found matching post but LinkURL %s was found in cache :\nsubreddit: %s\nby: %s",
				post.URL, post.Subreddit, post.Author)
			return nil
		}

		log.Infof("Found matching post:\nLink: %s\nsubreddit: %s\ntitle: %s\nSelf Text: %s\nby: %s",
			post.Permalink, post.Subreddit, post.Title, post.SelfText, post.Author)
		err := r.bot.Reply(post.Name, Text)
		if err != nil {
			log.Errorf("Got error trying to post reply: %v", err)
		}
	}

	return nil
}

func main() {

	// stop the cache janitor when we exit
	defer cache.stopJanitor()

	// run the cache janitor in a go routine
	go cache.janitor()

	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1,
		logging.MustStringFormatter(`%{color}%{time:Jan _2 15:04:05.000} %{level:.4s} ▶%{color:reset} %{message}`))

	fh, err := os.OpenFile("bot.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		fmt.Printf("Error opening log file: %v", err)
		return
	}

	backendFile := logging.NewLogBackend(fh, "", 0)
	backendFileFormatter := logging.NewBackendFormatter(backendFile,
		logging.MustStringFormatter(`%{time:Jan _2 15:04:05.000} %{level:.4s} ▶ %{message}`))

	logging.SetBackend(backend1Formatter, backendFileFormatter)

	log.Infof("Started suicide prevention bot.")

	if bot, err := reddit.NewBotFromAgentFile("bot.agent", 0); err != nil {
		log.Errorf("Failed to create bot handle: %v", err)
	} else {
		cfg := graw.Config{SubredditComments: []string{"all"},
			Subreddits: []string{"all"}}
		handler := &spBot{bot: bot}
		if _, wait, err := graw.Run(handler, bot, cfg); err != nil {
			log.Errorf("Failed to start graw run: %v", err)
		} else {
			errStr := fmt.Sprintf("%v", wait())

			if strings.Contains(errStr, "bad response") {
				log.Warningf("Restarting bot. Graw got a bad response error: %s", errStr)
				time.Sleep(time.Millisecond * 500)
				main()
			}

			if strings.Contains(errStr, "token expired") || strings.Contains(errStr, "connection reset") {
				log.Warningf("Restarting bot. Graw got a token error: %s", errStr)
				main()
			}

			log.Errorf("Graw run failed with error %s", errStr)
		}
	}
}
