package util

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/twodarek/go-twitter/twitter"

	_ "github.com/lib/pq"
)

func InsertFavoritedTweet(fulltext string, tweet twitter.Tweet, mediaType string, db *sql.DB) error {
	tweetJson, err := json.Marshal(tweet)
	if err != nil {
		fmt.Printf("Error while marshaling tweet %d, error: %s", tweet.ID, err)
		return err
	}
	res, err := db.Exec("INSERT INTO twitter_archive.public.favorites(tweet_fulltext, tweet_creator, tweet_json, media_type) VALUES ($1, $2, $3, $4)", fulltext, tweet.User.ScreenName, tweetJson, mediaType)
	if err != nil {
		fmt.Printf("Error while storing tweet %d to the database, error: %s", tweet.ID, err)
		return err
	}
	rowsEffected, err := res.RowsAffected()
	fmt.Sprintf("%d rows written!", rowsEffected)
	return err
}
