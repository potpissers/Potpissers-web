package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type redditVideoPost struct {
	YoutubeEmbedUrl string `json:"youtube_embed_url"`
	VideoUrl        string `json:"video_url"`
	PostUrl         string `json:"post_url"`
	Title           string `json:"title_url"`
}

var redditVideoPosts = []redditVideoPost{}

type redditImagePost struct {
	ImageUrl string `json:"image_url"`
	PostUrl  string `json:"post_url"`
}

var redditImagePosts = []redditImagePost{}

var imageRegex = regexp.MustCompile(`(?i)^(https?://)?(i\.redd\.it|i\.imgur\.com)/.*\.(png|jpg|jpeg)$`)
var youtubeVideoIdRegex = regexp.MustCompile(`[?&]v=([a-zA-Z0-9_-]{11})`)
var lastCheckedRedditPostId string
var lastCheckedRedditPostCreatedUtc float64
var redditPostsChannel = make(chan struct{}, 1)

func init() {
	println("reddit started")
	redditVideoPosts, redditImagePosts = getRedditPostData(potpissersRedditApiUrl)
	http.HandleFunc("/api/reddit", func(w http.ResponseWriter, r *http.Request) {
		handleRedditPostDataUpdate()
	})
	println("reddit done")
}

const redditAccessTokenDataFileName = "reddit_token_data.txt"

var redditAccessToken string
var redditAccessTokenExpiration = func() time.Time {
	csv, err := os.ReadFile(redditAccessTokenDataFileName)
	if err != nil {
		return time.Time{}
	}
	parts := strings.Split(string(csv), ",")
	redditAccessToken = parts[0]
	time, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		log.Fatal(err)
	}
	return time
}()

func handleRedditPostDataUpdate() {
	select {
	case redditPostsChannel <- struct{}{}:
		{
			newVideoPosts, newImagePosts := getRedditPostData(potpissersRedditApiUrl + "&before=" + lastCheckedRedditPostId) // holy fuck sorted by new -> before is newer, after is older
			for _, post := range newVideoPosts {
				redditVideoPosts = append([]redditVideoPost{post}, redditVideoPosts...)
			}
			for _, post := range newImagePosts {
				redditImagePosts = append([]redditImagePost{post}, redditImagePosts...)
			}
			if len(newVideoPosts) > 0 || len(newImagePosts) > 0 {
				home = getMainTemplateBytes("hub")
				mz = getMainTemplateBytes("mz")
				hcf = getMainTemplateBytes("hcf" + currentHcfServerName)

				handle := func(t any) {
					jsonData, err := json.Marshal(t)
					if err != nil {
						log.Fatal(err)
					}
					handleSseData(jsonData, mainConnections)
				}
				for _, post := range newVideoPosts {
					handle(sseMessage{"videos", post})
				}
				// TODO -> text posts
			}
			<-redditPostsChannel
		}
	default:
		return
	}
}
func getRedditPostData(redditApiUrl string) ([]redditVideoPost, []redditImagePost) {
	for redditAccessToken == "" || redditAccessTokenExpiration.Before(time.Now()) {
		println("started reddit api key")
		data := url.Values{}
		data.Set("grant_type", "client_credentials")
		req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("REDDIT_CLIENT_ID")+":"+os.Getenv("REDDIT_CLIENT_SECRET"))))
		client := &http.Client{}
		authResp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		//		println(authResp.StatusCode) // -> reddit access token rate limits hahd
		defer authResp.Body.Close()
		var result struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
			Scope       string `json:"scope"`
		}
		if err := json.NewDecoder(authResp.Body).Decode(&result); err != nil {
			log.Fatal(err)
		}
		redditAccessToken = result.AccessToken
		redditAccessTokenExpiration = time.Now().Add(time.Duration(result.ExpiresIn-45) * time.Second)
		err = os.WriteFile(redditAccessTokenDataFileName, []byte(redditAccessToken+","+redditAccessTokenExpiration.Format(time.RFC3339)), 0644)
		if err != nil {
			log.Fatal(err)
		}
		println("retrieved reddit api key")
	}

	var videoPosts []redditVideoPost
	var imagePosts []redditImagePost
	for {
		req, err := http.NewRequest("GET", redditApiUrl, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+redditAccessToken)
		println("reddit request")
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		var responseJson struct {
			Kind string `json:"kind"`
			Data struct {
				After     *string `json:"after"`
				Dist      int     `json:"dist"`
				Modhash   string  `json:"modhash"`
				GeoFilter string  `json:"geo_filter"`
				Children  []struct {
					Kind string `json:"kind"`
					Data struct {
						Name        string  `json:"name"`
						Subreddit   string  `json:"subreddit"`
						Title       string  `json:"title"`
						Selftext    string  `json:"selftext"`
						Author      string  `json:"author"`
						UpvoteRatio float64 `json:"upvote_ratio"`
						Thumbnail   string  `json:"thumbnail"`
						URL         string  `json:"url"`
						NumComments int     `json:"num_comments"`
						Permalink   string  `json:"permalink"`
						CreatedUTC  float64 `json:"created_utc"`
						IsVideo     bool    `json:"is_video"`
						Media       *struct {
							RedditVideo *struct {
								FallbackURL  string `json:"fallback_url"`
								Height       int    `json:"height"`
								Width        int    `json:"width"`
								Duration     int    `json:"duration"`
								ThumbnailURL string `json:"thumbnail_url"`
							} `json:"reddit_video"`
						} `json:"media,omitempty"`
					} `json:"data"`
				} `json:"children"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&responseJson)
		if err != nil {
			println(err.Error())
			continue
		}

		children := responseJson.Data.Children
		if len(children) > 0 {
			if lastCheckedRedditPostCreatedUtc < children[0].Data.CreatedUTC {
				lastCheckedRedditPostId = children[0].Data.Name
				lastCheckedRedditPostCreatedUtc = children[0].Data.CreatedUTC
			}
			for _, child := range children {
				getRedditPostUrl := func(permalink string) string {
					return "https://www.reddit.com" + permalink
				}

				data := child.Data
				linkPostUrl := data.URL

				if imageRegex.MatchString(linkPostUrl) {
					imagePosts = append(imagePosts, redditImagePost{linkPostUrl, getRedditPostUrl(data.Permalink)})
				} else if strings.Contains(linkPostUrl, "youtube.com") || strings.Contains(linkPostUrl, "youtu.be") { // TODO https ?
					videoPosts = append(videoPosts, redditVideoPost{
						YoutubeEmbedUrl: "https://www.youtube.com/embed/" + youtubeVideoIdRegex.FindStringSubmatch(data.URL)[1],
						PostUrl:         getRedditPostUrl(data.Permalink),
						Title:           data.Title,
					})
				} else if data.Media != nil {
					videoPosts = append(videoPosts, redditVideoPost{
						VideoUrl: data.URL,
						PostUrl:  getRedditPostUrl(data.Permalink),
						Title:    data.Title,
					})
				}
			}
		}
		break
	}
	return videoPosts, imagePosts
}
