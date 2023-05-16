package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

var cache map[string]*CacheItem

type Body struct {
	Token string `json:"token" binding:"required"`
	URL   string `json:"url" binding:"required"`
}

type CacheItem struct {
	URL    *url.URL
	Client *http.Client
}

func main() {
	cache = make(map[string]*CacheItem, 0)
	router := gin.New()
	router.POST("/test", func(context *gin.Context) {
		body := Body{}
		if err := context.BindJSON(&body); err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
			return
		}

		jar, err := cookiejar.New(nil)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}

		client := &http.Client{
			Jar: jar,
		}

		resp, err := client.Get(body.URL + "&k8s_token=" + body.Token)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}

		gh_url, err := extractAuthURI(resp.Body)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}

		url, err := replaceRedirectURIAndCacheResult(gh_url, "http://localhost:8080/oauth/callback", client)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}

		bu, err := url.Parse(body.URL + "&k8s_token=" + body.Token)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}
		fmt.Println("jar has these cookies after first call: ", client.Jar.Cookies(bu))

		context.JSON(http.StatusAccepted, url.String())
	})

	router.GET("/oauth/callback", func(context *gin.Context) {
		values := context.Request.URL.Query()
		state := values.Get("state")

		client := cache[state].Client
		uri := cache[state].URL
		uri.RawQuery = context.Request.URL.RawQuery
		fmt.Println("redirecting to spi to: ", uri.String())
		resp, err := client.Get(uri.String())
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}
		fmt.Println("jar has these cookies after first call: ", client.Jar.Cookies(uri))
		context.JSON(http.StatusAccepted, string(body))
	})
	router.Run(":8080")
}

func extractAuthURI(resp io.Reader) (*url.URL, error) {
	z := html.NewTokenizer(resp)

	for {
		tt := z.Next()

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "meta" {
				for _, attr := range t.Attr {
					if attr.Key == "http-equiv" && attr.Val == "refresh" {
						for _, a := range t.Attr {
							if a.Key == "content" && strings.Contains(a.Val, "url=") {
								uri := strings.Split(a.Val, "url=")
								u, err := url.Parse(uri[1])
								if err != nil {
									return nil, err
								}
								return u, nil
							}
						}
					}
				}
			}
		case html.ErrorToken:
			return nil, fmt.Errorf("no URI found")
		}

	}
}

func replaceRedirectURIAndCacheResult(u *url.URL, with string, cli *http.Client) (*url.URL, error) {

	values := u.Query()
	state := values.Get("state")
	redirect_uri, err := url.Parse(values.Get("redirect_uri"))
	if err != nil {
		return nil, fmt.Errorf("cannot parse redirect_uri from orginal url")
	}

	cache[state] = &CacheItem{}
	cache[state].URL = redirect_uri
	cache[state].Client = cli
	values.Set("redirect_uri", with)
	u.RawQuery = values.Encode()

	return u, nil
}
