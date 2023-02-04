package main

import (
	"fmt"
	"github.com/dghubble/oauth1"
	"github.com/twodarek/go-twitter/twitter"
)

func main() {
	fmt.Printf("Starting download")

	config := oauth1.NewConfig("consumerKey", "consumerSecret")
	token := oauth1.NewToken("accessToken", "accessSecret")
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	truePtr := true
	listParams := twitter.FavoriteListParams{
		ScreenName:      "thomaswodarek",
		Count:           1,
		IncludeEntities: &truePtr,
	}
	favorites, res, err := client.Favorites.List(&listParams)
	if err != nil || res.StatusCode > 299 {
		fmt.Printf("Unable to get list of favorite tweets.  Error: %s", err)
	}

	usersInfo := map[string]twitter.User{}
	for _, fav := range favorites {
		username := fav.User.ScreenName
		if _, exists := usersInfo[username]; !exists {
			usersInfo[username] = *fav.User
		}
	}

	tweetContent := favorites[0].Text
	tweetCreator := favorites[0].User.ScreenName
	fmt.Printf("liked tweet: %s by %s", tweetContent, tweetCreator)

}
