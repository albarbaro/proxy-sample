package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

var cache map[string]*CacheItem
var client *http.Client

type Body struct {
	Token string `json:"token" form:"token" binding:"required"`
	URL   string `json:"url" form:"url" binding:"required"`
}

type CacheItem struct {
	URL    *url.URL
	Client *http.Client
}

func main() {
	cache = make(map[string]*CacheItem, 0)
	router := gin.New()

	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	client = &http.Client{
		Jar: jar,
	}

	router.POST("/test", func(context *gin.Context) {
		body := Body{}
		if err := context.Bind(&body); err != nil {
			fmt.Println(err)
			context.AbortWithError(http.StatusBadRequest, err)
			return
		}

		req, _ := http.NewRequest("GET", body.URL+"&k8s_token="+body.Token, nil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
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

		_, err = url.Parse(body.URL + "&k8s_token=" + body.Token)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}

		context.Redirect(http.StatusFound, url.String())

	})

	router.GET("/oauth/callback", func(context *gin.Context) {
		values := context.Request.URL.Query()
		state := values.Get("state")

		uri := cache[state].URL
		uri.RawQuery = context.Request.URL.RawQuery

		req, _ := http.NewRequest("GET", uri.String(), nil)
		req.Header.Set("Content-Type", "application/json")
		for _, cookie := range client.Jar.Cookies(uri) {
			http.SetCookie(context.Writer, cookie)
		}

		resp, err := client.Do(req)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
		}

		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("REQUEST:\n%s", string(reqDump))

		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("RESPONSE:\n%s\n", string(respDump))

		body, error := ioutil.ReadAll(resp.Body)
		if error != nil {
			fmt.Println(error)
		}
		resp.Body.Close()

		context.JSON(http.StatusAccepted, string(body))
	})

	router.GET("/index", func(context *gin.Context) {
		markdown := `<!DOCTYPE html><html><body><h2>HTML Forms</h2><form action="http://localhost:8080/test" method="POST"> <label for="fname">url:</label><br> <input type="text" id="url" name="url" value=""><br> <label for="lname">token:</label><br> <input type="text" id="token" name="token" value=""><br><br> <input type="submit" value="Submit"></form> </body></html>`
		context.Data(http.StatusOK, "text/html; charset=utf-8", []byte(markdown))
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
