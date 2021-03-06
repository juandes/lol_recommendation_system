package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/juandes/lol-recommendation-system/data"
	"github.com/juandes/lol-recommendation-system/recommender"
	"github.com/juandes/lol-recommendation-system/vectormath"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	trainingData = pflag.String("trainingset", "../../static/winning_teams.csv", "path to training dataset")
	// Had to read the PORT from an env variable because of Heroku
	port = pflag.String("port", os.Getenv("PORT"), "listening port")

	championToIndex = make(map[string]int)
	indexToChampion = make(map[int]string)
)

// PredictionInput is the object to which the inputted JSON will be unmarshaled to
type PredictionInput struct {
	Champions   []string `json:"champions"`
	Intercept   bool     `json:"intercept"`
	Shuffle     bool     `json:"shuffle"`
	Serendipity bool     `json:"serendipity"`
}

// Output is the return response that contains the recommendations
type Output struct {
	Recommendations [][]string `json:"recommendations"`
}

// TODO (Juan): add an error structure. See this: https://blog.restcase.com/rest-api-error-codes-101/

// curl -d '{"key1":"value1", "key2":"value2"}' -H "Content-Type: application/json" -X POST http://localhost:8080/recommend
// curl -d '{"champions":["jax", "ashe", "drmundo"], "intercept": true, "shuffle": true}' -H "Content-Type: application/json" -X POST http://localhost:8080/recommend

func main() {
	pflag.Parse()
	if *port == "" {
		*port = "8080"
	}

	// read the training set
	train, header, err := data.ReadData(*trainingData)
	if err != nil {
		log.Fatalf("Error reading training set: %v", err)
	}

	for i, val := range header {
		name := strings.ToLower(val)
		championToIndex[name] = i
		indexToChampion[i] = name
	}

	log.Info("Starting recommendation service...")
	log.Infof("Listening on port %s", *port)

	// create the recommender engine with k = 5
	engine := recommender.NewNeighborhoodBasedRecommender(train, 5)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/recommend", recommendationHandler(engine))

	go log.Fatal(http.ListenAndServe(":"+*port, mux))
}

func recommendationHandler(engine *recommender.NeighborhoodBasedRecommender) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pinput PredictionInput
		input := make([]float64, len(championToIndex))

		if r.Method != "POST" {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}

		err = json.Unmarshal(body, &pinput)
		if err != nil {
			// TODO: Replace logs for http.Error
			log.Errorf("Error unmarshaling: %v", err)
			return
		}

		if len(pinput.Champions) == 0 || len(pinput.Champions) > 4 {
			log.Error("Invalid number of champions. Please provide 4 or less")
			return
		}

		// create the feature vector by writing 1
		// in the index corresponding to the champion
		for _, val := range pinput.Champions {
			item, ok := championToIndex[val]
			if !ok {
				log.Warningf("Unknown champion: %v", val)
				return
			}
			input[item] = 1
		}

		recommendations, err := engine.Recommend(input, vectormath.Pearson, pinput.Intercept, pinput.Shuffle, pinput.Serendipity)
		if err != nil {
			log.Errorf("Error predicting recommendation: %v", err)
			return
		}

		allRecommendations := [][]string{}
		for _, val := range recommendations {
			recommendedItem := []string{}
			for i, isRecommended := range val.GetRecommendation() {
				if isRecommended != 1 {
					continue
				}
				champ := indexToChampion[int(i)]
				recommendedItem = append(recommendedItem, champ)
			}
			allRecommendations = append(allRecommendations, recommendedItem)
		}

		response, _ := json.Marshal(&Output{
			Recommendations: allRecommendations,
		})
		fmt.Fprintln(w, string(response))
	})
}
