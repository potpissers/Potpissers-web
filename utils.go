package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func handleFatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func handleFatalPgx(_ pgconn.CommandTag, err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func handleGetFatalJsonT[T any](request *http.Request) T {
	resp, err := (&http.Client{}).Do(request)
	handleFatalErr(err)
	defer resp.Body.Close()
	return getFatalJsonT[T](resp)
}

func getFatalJsonT[T any](resp *http.Response) T {
	var messages T
	handleFatalErr(json.NewDecoder(resp.Body).Decode(&messages))
	return messages
}

type sseMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func handleSseData(sseConnections *[]sseConnection, bytes []byte, additionalConnections ...*[]sseConnection) {
	for i := range additionalConnections {
		go handleSseData(additionalConnections[i], bytes)
	}

	validConnChan := make(chan sseConnection, len(*sseConnections))
	var waitGroup sync.WaitGroup
	for _, c := range *sseConnections {
		waitGroup.Add(1)
		go func(conn sseConnection) {
			defer waitGroup.Done()

			conn.mutex.Lock()
			if _, err := conn.response.Write(bytes); err == nil {
				conn.flusher.Flush()
				validConnChan <- conn
			}
			conn.mutex.Unlock()
		}(c)
	}
	waitGroup.Wait()

	close(validConnChan)

	var validConnections []sseConnection
	for conn := range validConnChan {
		validConnections = append(validConnections, conn)
	}
	*sseConnections = validConnections // TODO -> atomic pointer could be used here pog
}

func getMojangApiUuidRequest(username string) (*http.Response, error) {
	return http.Get(minecraftUsernameLookupUrl + username)
}

func getRowsBlocking(query string, bar func(rows pgx.Rows), params ...any) {
	rows, err := postgresPool.Query(context.Background(), query, params...)
	defer rows.Close()
	handleFatalErr(err)
	bar(rows)
}

func getFatalRequest(method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	handleFatalErr(err)
	return req
}
func addSquareHeaders(request *http.Request) {
	request.Header.Add("Authorization", "Bearer "+os.Getenv("SQUARE_ACCESS_TOKEN"))
	request.Header.Add("Content-Type", "application/json")
}

var redditAccessToken string
var redditAccessTokenExpiration time.Time
func getRedditPostData(redditApiUrl string) ([]redditVideoPost, []redditImagePost, string) {
	for redditAccessToken == "" || redditAccessTokenExpiration.Before(time.Now()) {
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
		redditAccessTokenExpiration = time.Now().Add(time.Duration(result.ExpiresIn - 45) * time.Second)
	}

	req, err := http.NewRequest("GET", redditApiUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+redditAccessToken)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	responseJson := getFatalJsonT[struct {
		Kind string `json:"kind"`
		Data struct {
			After     *string `json:"after"`
			Dist      int     `json:"dist"`
			Modhash   string  `json:"modhash"`
			GeoFilter string  `json:"geo_filter"`
			Children  []struct {
				Kind string `json:"kind"`
				Data struct {
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
	}](resp)

	var videoPosts []redditVideoPost
	var imagePosts []redditImagePost
	children := responseJson.Data.Children
	for _, child := range children {
		getRedditPostUrl := func(permalink string) string {
			return "https://www.reddit.com" + permalink
		}

		data := child.Data
		linkPostUrl := data.URL

		if imageRegex.MatchString(linkPostUrl) {
			imagePosts = append(imagePosts, redditImagePost{linkPostUrl, getRedditPostUrl(data.Permalink)})
		} else if strings.HasPrefix(linkPostUrl, "https://youtube.com") || strings.HasPrefix(linkPostUrl, "https://youtu.be") {
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
	return videoPosts, imagePosts, redditPostIdRegex.FindStringSubmatch(children[0].Data.Permalink)[1]
}

func handleRedditPostDataUpdate() {
	select {
	case redditPostsChannel <- struct{}{}:
		{
			var newVideoPosts []redditVideoPost
			var newImagePosts []redditImagePost
			newVideoPosts, newImagePosts, lastCheckedRedditPostId = getRedditPostData(potpissersRedditApiUrl + "&after=" + lastCheckedRedditPostId)
			for _, post := range newVideoPosts {
				redditVideoPosts = append([]redditVideoPost{post}, redditVideoPosts...)
			}
			for _, post := range newImagePosts {
				redditImagePosts = append([]redditImagePost{post}, redditImagePosts...)
			}
			<-redditPostsChannel

			handle := func(t any) {
				jsonData, err := json.Marshal(t)
				if err != nil {
					log.Fatal(err)
				}
				handleSseData(&homeConnections, jsonData, &hcfConnections, &mzConnections)
			}
			for _, post := range newVideoPosts {
				handle(sseMessage{"videos", post})
			}
			// TODO -> text posts
		}
	default:
		return
	}
}

func handleDiscordMessagesUpdate(channel chan struct{}, discordChannelId string, mostRecentMessageId *string, slice *[]discordMessage, sseMessageType string) {
	select {
	case channel <- struct{}{}:
		{
			newMessages := getDiscordMessages(discordChannelId, "after="+*mostRecentMessageId+"&")
			if len(messages) > 0 {
				*mostRecentMessageId = newMessages[0].ID

				for _, msg := range newMessages {
					*slice = append([]discordMessage{msg}, *slice...)

					jsonData, err := json.Marshal(sseMessage{sseMessageType, msg})
					if err != nil {
						log.Fatal(err)
					}
					handleSseData(&homeConnections, jsonData, &mzConnections, &hcfConnections)
				}
			}
			<-channel
		}
	default:
		return
	}
}
