package main

import (
  "net/http"
  "io/ioutil"
	"encoding/json"
  "fmt"
	"net/url"
	"time"

  "github.com/gin-gonic/contrib/static"
  "github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

type Gender string

var Genders = struct {
	Unspecified Gender
	Male        Gender
	Female      Gender
}{"", "M", "F"}

type AthleteMeta struct {
	Id int64 `json:"id"`
}

type AthleteSummary struct {
	AthleteMeta
	FirstName        string    `json:"firstname"`
	LastName         string    `json:"lastname"`
	ProfileMedium    string    `json:"profile_medium"` // URL to a 62x62 pixel profile picture
	Profile          string    `json:"profile"`        // URL to a 124x124 pixel profile picture
	City             string    `json:"city"`
	State            string    `json:"state"`
	Country          string    `json:"country"`
	Gender           Gender    `json:"sex"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AthleteDetailed struct {
	AthleteSummary
	Email                 string         `json:"email"`
	DatePreference        string         `json:"date_preference"`
	MeasurementPreference string         `json:"measurement_preference"`
}

type AuthorizationResponse struct {
	AccessToken string          `json:"access_token"`
	State       string          `json:"State"`
	Athlete     AthleteDetailed `json:"athlete"`
}

func main() {

	creds := Creds()

	// Set the router as the default one shipped with Gin
	router := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

  // Serve frontend static files
  router.Use(static.Serve("/", static.LocalFile("./views", true)))

	// strava login
  router.GET("/login", func(c *gin.Context) {
    c.Redirect(http.StatusFound, fmt.Sprintf("http://www.strava.com/oauth/authorize?client_id=%d&response_type=code&redirect_uri=http://localhost:3000/callback&approval_prompt=force&scope=activity:read_all,profile:read_all", creds.client_id))
  })

  router.GET("/callback", func(c *gin.Context) {
    // state := c.Query("state")
    // code := c.Query("code")
    // scope := c.Query("scope")

    resp, err := http.PostForm("https://www.strava.com/oauth/token",
		url.Values{"client_id": {fmt.Sprintf("%d", creds.client_id)}, "client_secret": {creds.client_secret}, "code": {c.Query("code")}})
    if err != nil {
        print(err)
    }
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
        print(err)
    }
		var result AuthorizationResponse
		json.Unmarshal(body, &result)
		// fmt.Printf("%+v", result.Athlete)

		session := sessions.Default(c)
		session.Set("firstname", result.Athlete.FirstName)
		session.Set("token", result.AccessToken)
		session.Save()

		/*v := session.Get("firstname")
		if v == nil {
			count = ""
		} else {
			count = v.(string)
			count++
		}
		session.Set("count", count)*/

    c.Redirect(http.StatusFound, "/dashboard")
  })

  router.GET("/dashboard", func(c *gin.Context) {
		session := sessions.Default(c)
		name := session.Get("firstname")
		token := session.Get("token")

		if token == nil {
			c.Redirect(http.StatusFound, "/login")
		}

		c.JSON(http.StatusOK, gin.H {
      "firstName": name,
    })
	})

  // Setup route group for the API
  api := router.Group("/api")
  {
    api.GET("/", func(c *gin.Context) {
      c.JSON(http.StatusOK, gin.H {
        "message": "pong",
      })
    })

    api.GET("/activities", func(c *gin.Context) {

			session := sessions.Default(c)
			token := session.Get("token")

			t := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
			url := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?after=%d&access_token=%s", t, token)
			fmt.Println(url)

			if token == nil {
				c.Redirect(http.StatusFound, "/login")
			}

			resp, err := http.Get(url)
			if err != nil {
        print(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
        print(err)
			}


			fmt.Print(string(body))

      c.JSON(http.StatusOK, gin.H {
        "message": "pong",
      })
    })


	}

  // Start and run the server
  router.Run(":3000")
}
