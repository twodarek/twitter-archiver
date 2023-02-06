package main

import (
	"database/sql"
	"fmt"
	"github.com/dghubble/oauth1"
	"github.com/twodarek/go-twitter/twitter"
	"github.com/twodarek/twitter-archiver/util"

	_ "github.com/lib/pq"
	config2 "github.com/twodarek/twitter-archiver/config"
)

func main() {
	fmt.Printf("Starting download")

	appConfig, err := config2.LoadConfig("./config")

	config := oauth1.NewConfig(appConfig.ConsumerKey, appConfig.ConsumerSecret)
	token := oauth1.NewToken(appConfig.AccessToken, appConfig.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	//PostGreSQL client
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		appConfig.DatabaseHost, appConfig.DatabasePort, appConfig.DatabaseUser, appConfig.DatabasePass, appConfig.DatabaseName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Printf("Error connecting to database: %s", err)
		//os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Printf("Error pinging to database: %s", err)
		//os.Exit(1)
	}

	fmt.Println("Successfully connected to database!")

	// This _should_ be larger than the most recent tweet favorited (so that we get the most recent favorited tweets
	nextRunMaxID := int64(2000000000000000000)
	stillRunning := 1
	tmpLoopControl := 0

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
		if err != nil || res.StatusCode > 299 {
			fmt.Printf("Unable to get list of favorite tweets.  Error: %s", err)
		}

		if len(favorites) < 1 || tmpLoopControl > 2 {
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

			if fav.Truncated {
				fmt.Printf("favorite %d, author %s, content: %s\n", i, fav.User.ScreenName, fav.RetweetedStatus.FullText)
			} else {
				fmt.Printf("favorite %d, author: %s, content: %s\n", i, fav.User.ScreenName, fav.FullText)
			}

			if fav.ExtendedEntities != nil {
				for j, media := range fav.ExtendedEntities.Media {
					downloadableMediaUrl := ""
					switch media.Type {
					case util.TwitterMediaTypePhoto:
						downloadableMediaUrl = media.MediaURLHttps
					case util.TwitterMediaTypeVideo, util.TwitterMediaTypeAnimatedGif:
						downloadableMediaUrl = util.GetHighestBitrateVariant(media.VideoInfo.Variants).URL
					}

					fmt.Printf("MEDIA FOUND!! idx: %d, type: %s, url: %s\n", j, media.Type, downloadableMediaUrl)
				}
			}
			fmt.Printf("\n\n\n\n")
		}
		tmpLoopControl++
		fmt.Printf("LOOP: %d\n", tmpLoopControl)
	}
}
