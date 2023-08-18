package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Train struct {
	TrainName      string `json:"trainName"`
	TrainNumber    string `json:"trainNumber"`
	DepartureTime  Time   `json:"departureTime"`
	SeatsAvailable Seats  `json:"seatsAvailable"`
	Price          Prices `json:"price"`
	DelayedBy      int    `json:"delayedBy"`
}

type Time struct {
	Hours   int `json:"Hours"`
	Minutes int `json:"Minutes"`
	Seconds int `json:"Seconds"`
}

type Seats struct {
	Sleeper int `json:"sleeper"`
	AC      int `json:"AC"`
}

type Prices struct {
	Sleeper int `json:"sleeper"`
	AC      int `json:"AC"`
}

type AuthResponse struct {
	Token     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
}

var authToken string
var authTokenExpiration time.Time

// CORSMiddleware adds CORS headers to requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isAuthTokenValid checks if the authentication token is valid
func isAuthTokenValid() bool {
	return authToken != "" && authTokenExpiration.After(time.Now())
}

// fetchAuthToken fetches a new authentication token from the API
func fetchAuthToken() error {
	authURL := os.Getenv("AUTH_TOKEN_URL")
	clientID := os.Getenv("clientID")
	companyName := os.Getenv("companyName")
	ownerName := os.Getenv("ownerName")
	ownerEmail := os.Getenv("ownerEmail")
	rollNo := os.Getenv("rollNo")
	clientSecret := os.Getenv("clientSecret")
	payload := map[string]string{
		"companyName":  companyName,
		"clientID":     clientID,
		"clientSecret": clientSecret,
		"ownerName":    ownerName,
		"ownerEmail":   ownerEmail,
		"rollNo":       rollNo,
	}

	payloadBytes, _ := json.Marshal(payload)

	response, err := http.Post(authURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var authResponse AuthResponse
	if err := json.NewDecoder(response.Body).Decode(&authResponse); err != nil {
		return err
	}
	authToken = authResponse.Token
	authTokenExpiration = time.Unix(authResponse.ExpiresIn, 0)

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	localServerUrl := os.Getenv("LOCAL_SERVER_BASE_URL")

	router := gin.Default()
	router.Use(CORSMiddleware())

	router.GET(localServerUrl, getAllTrains)

	router.Run(":8080")
}

// get all trains from the API
func getAllTrains(c *gin.Context) {
	if !isAuthTokenValid() {
		if err := fetchAuthToken(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch authentication token"})
			return
		}
	}

	getTrainsUrl := os.Getenv("TRAIN_URL")

	req, err := http.NewRequest("GET", getTrainsUrl, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch train data"})
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var trainData []Train
	if err := json.Unmarshal(responseBody, &trainData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse train data"})
		return
	}

	// Filter out trains departing in the next 30 minutes
	currentTime := time.Now()
	filteredTrains := []Train{}
	for _, train := range trainData {
		departureTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), train.DepartureTime.Hours, train.DepartureTime.Minutes, train.DepartureTime.Seconds, 0, time.UTC)
		if departureTime.Sub(currentTime) > 30*time.Minute {
			filteredTrains = append(filteredTrains, train)
		}
	}

	// Sort trains based on specified criteria
	sort.SliceStable(filteredTrains, func(i, j int) bool {
		if filteredTrains[i].Price.Sleeper != filteredTrains[j].Price.Sleeper {
			return filteredTrains[i].Price.Sleeper < filteredTrains[j].Price.Sleeper
		}
		if filteredTrains[i].SeatsAvailable.Sleeper != filteredTrains[j].SeatsAvailable.Sleeper {
			return filteredTrains[i].SeatsAvailable.Sleeper > filteredTrains[j].SeatsAvailable.Sleeper
		}
		departureTimeI := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), filteredTrains[i].DepartureTime.Hours, filteredTrains[i].DepartureTime.Minutes, filteredTrains[i].DepartureTime.Seconds, 0, time.UTC)
		departureTimeJ := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), filteredTrains[j].DepartureTime.Hours, filteredTrains[j].DepartureTime.Minutes, filteredTrains[j].DepartureTime.Seconds, 0, time.UTC)
		departureTimeI = departureTimeI.Add(time.Duration(filteredTrains[i].DelayedBy) * time.Minute)
		departureTimeJ = departureTimeJ.Add(time.Duration(filteredTrains[j].DelayedBy) * time.Minute)
		return departureTimeI.After(departureTimeJ)
	})

	c.JSON(http.StatusOK, filteredTrains)
}
