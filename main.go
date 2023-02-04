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
		Count:           200,
		IncludeEntities: &truePtr,
	}
	favorites, res, err := client.Favorites.List(&listParams)
	if err != nil || res.StatusCode > 299 {
		fmt.Printf("Unable to get list of favorite tweets.  Error: %s", err)
	}

	usersToPullInfoFor := []twitter.User{}
	for _, fav := range favorites {
		usersToPullInfoFor = append(usersToPullInfoFor, *fav.User)
	}

}
