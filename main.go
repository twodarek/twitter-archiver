package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dghubble/oauth1"
	"github.com/twodarek/go-twitter/twitter"
	"github.com/twodarek/twitter-archiver/util"
	"net/http"
	"strconv"
	"strings"
	"time"

	v2 "github.com/g8rswimmer/go-twitter/v2"
	_ "github.com/lib/pq"
	config2 "github.com/twodarek/twitter-archiver/config"
)

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

func main() {
	fmt.Printf("Starting download")

	appConfig, err := config2.LoadConfig("./config")

	config := oauth1.NewConfig(appConfig.ConsumerKey, appConfig.ConsumerSecret)
	token := oauth1.NewToken(appConfig.AccessToken, appConfig.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)
	v2client := &v2.Client{
		Authorizer: authorize{
			Token: appConfig.APIBearerToken,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}

	//PostGreSQL client
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		appConfig.DatabaseHost, appConfig.DatabasePort, appConfig.DatabaseUser, appConfig.DatabasePass, appConfig.DatabaseName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Printf("Error connecting to database: %s", err.Error())
		//os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Printf("Error pinging to database: %s", err.Error())
		//os.Exit(1)
	}

	fmt.Println("Successfully connected to database!")

	// This _should_ be larger than the most recent tweet favorited (so that we get the most recent favorited tweets
	nextRunMaxID := int64(2000000000000000000)
	stillRunning := 1
	tmpLoopControl := 0
	tweetExpansions := []v2.Expansion{v2.ExpansionAuthorID, v2.ExpansionReferencedTweetsID, v2.ExpansionInReplyToUserID, v2.ExpansionAttachmentsMediaKeys, v2.ExpansionEntitiesMentionsUserName, v2.ExpansionReferencedTweetsIDAuthorID}

	tweetIDs := []string{}

	truePtr := true
	for stillRunning >= 1 {
		listParams := twitter.FavoriteListParams{
			ScreenName:      "thomaswodarek",
			Count:           10,
			IncludeEntities: &truePtr,
			TweetMode:       "extended",
			MaxID:           nextRunMaxID,
		}
		favorites, res, err := client.Favorites.List(&listParams)
		if err != nil {
			if res.StatusCode > 299 && res.StatusCode < 500 {
				fmt.Printf("Unable to get list of favorite tweets.  Error: %s", err)
			}
		}

		if len(favorites) < 1 || tmpLoopControl > 0 {
			stillRunning = 0
		}

		usersInfo := map[string]twitter.User{}
		for _, fav := range favorites {
			username := fav.User.ScreenName
			if _, exists := usersInfo[username]; !exists {
				usersInfo[username] = *fav.User
			}
		}

		for i, fav := range favorites {
			if fav.ID < nextRunMaxID {
				nextRunMaxID = fav.ID
			}
			tweetIDs = append(tweetIDs, strconv.FormatInt(fav.ID, 10))

			fulltext := ""
			if fav.Truncated {
				fmt.Printf("favorite %d, author %s, content: %s\n", i, fav.User.ScreenName, fav.RetweetedStatus.FullText)
				fulltext = fav.RetweetedStatus.FullText
			} else {
				fmt.Printf("favorite %d, author: %s, content: %s\n", i, fav.User.ScreenName, fav.FullText)
				fulltext = fav.FullText
			}

			mediaType := ""
			if fav.ExtendedEntities != nil {
				for j, media := range fav.ExtendedEntities.Media {
					downloadableMediaUrl := ""
					switch media.Type {
					case util.TwitterMediaTypePhoto:
						downloadableMediaUrl = media.MediaURLHttps
					case util.TwitterMediaTypeVideo, util.TwitterMediaTypeAnimatedGif:
						downloadableMediaUrl = util.GetHighestBitrateVariant(media.VideoInfo.Variants).URL
					}
					mediaType = media.Type

					fmt.Printf("MEDIA FOUND!! idx: %d, type: %s, url: %s\n", j, media.Type, downloadableMediaUrl)
				}
			}

			// tweet_fulltext, tweet_creator, tweet_json, media_path, media_type
			err = util.InsertFavoritedTweet(fulltext, fav, mediaType, db)
			if err != nil {
				fmt.Sprintf("Error attempting to store tweet %d to the database.  Error: %s", fav.ID, err)
			}

			fmt.Printf("\n\n\n\n")
		}
		tmpLoopControl++
		fmt.Printf("LOOP: %d\n\n\n\n", tmpLoopControl)
	}

	conversationLookupTweetRes, err := v2client.ListTweetLookup(context.Background(), strings.Join(tweetIDs, ","), v2.ListTweetLookupOpts{
		Expansions:      tweetExpansions,
		TweetFields:     []v2.TweetField{v2.TweetFieldID, v2.TweetFieldConversationID},
		MaxResults:      200,
		PaginationToken: "",
	})
	if err != nil {
		fmt.Sprintf("Unable to lookup conversation IDs from the twitter api.  Error: %s", err)
	}

	conversationIDsToPull := []string{}
	for _, tweet := range conversationLookupTweetRes.Raw.Tweets {
		conversationIDsToPull = append(conversationIDsToPull, tweet.ConversationID)
	}

	searchOpts := v2.TweetSearchOpts{
		Expansions:  nil,
		MediaFields: nil,
		PlaceFields: nil,
		PollFields:  nil,
		TweetFields: nil,
		UserFields:  nil,
		StartTime:   time.Time{},
		EndTime:     time.Time{},
		SortOrder:   "",
		MaxResults:  0,
		NextToken:   "",
		SinceID:     "",
		UntilID:     "",
	}
	tweetIDsFromConversations := []string{}
	for _, conversationID := range conversationIDsToPull {
		conversationTweets, err := v2client.TweetSearch(context.Background(), fmt.Sprintf("conversation_id:%s", conversationID), searchOpts)
		if err != nil {
			fmt.Sprintf("Unable to find tweets by conversation ID %s, Error: %s", conversationID, err)
		}
		for _, tweet := range conversationTweets.Raw.Tweets {
			tweetIDsFromConversations = append(tweetIDsFromConversations, tweet.ID)
		}
	}

	//TODO(twodarek): get actual tweet objects from the v1.1 api that were found to be in a thread
}
