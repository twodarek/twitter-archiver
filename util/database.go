package util

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/twodarek/go-twitter/twitter"

	_ "github.com/lib/pq"
)

func GetLowestTweetId(db *sql.DB) (string, error) {
	var minTweetId string
	if err := db.QueryRow("SELECT (tweet_id) from twitter_archive.public.favorites").Scan(&minTweetId); err != nil {
		if err == sql.ErrNoRows {
			return "2000000000000000000", fmt.Errorf("Unable to get min tweet id, no records found, %s", err)
		}
		return "2000000000000000000", fmt.Errorf("Unable to get min tweet id, error: %s", err)
	}
	return minTweetId, nil
}

func InsertFavoritedTweet(fulltext string, tweet twitter.Tweet, mediaType string, db *sql.DB) error {
	tweetJson, err := json.Marshal(tweet)
	if err != nil {
		fmt.Printf("Error while marshaling tweet %d, error: %s", tweet.ID, err)
		return err
	}
	res, err := db.Exec("INSERT INTO twitter_archive.public.favorites(tweet_fulltext, tweet_creator, tweet_json, media_type, tweet_id) VALUES ($1, $2, $3, $4, $5)", fulltext, tweet.User.ScreenName, tweetJson, mediaType, tweet.ID)
	if err != nil {
		fmt.Printf("Error while storing tweet %d to the database, error: %s", tweet.ID, err)
		return err
	}
	rowsEffected, err := res.RowsAffected()
	fmt.Sprintf("%d rows written!", rowsEffected)
	return err
}

func InsertConversationID(conversationID, tweetID string, db *sql.DB) error {
	res, err := db.Exec("INSERT INTO twitter_archive.public.conversations(conversation_id, tweet_id, resolved) VALUES ($1, $2, $3)", conversationID, tweetID, true)
	if err != nil {
		fmt.Printf("Error while storing conversation %d to the database, error: %s", conversationID, err)
		return err
	}
	rowsEffected, err := res.RowsAffected()
	fmt.Sprintf("%d rows written!", rowsEffected)
	return err
}
