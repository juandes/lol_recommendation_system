package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/juandes/knn-recommender-system/recommender"
	vm "github.com/juandes/knn-recommender-system/vectormath"
)

func main() {
	// one hot encoding of the different items
	// 1 means that the "user" has liked, bought, seen ...
	// the item. 0 means the user has not seen the item.
	data := [][]float64{
		[]float64{1.0, 1.0, 1.0, 0.0, 1.0, 0.0},
		[]float64{1.0, 1.0, 0.0, 0.0, 1.0, 0.0},
		// Arrays of only 0's causes weird behaviour, do not use them :)
		//[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 1.0, 0.0},
		[]float64{0.0, 1.0, 0.0, 0.0, 1.0, 0.0},
	}

	reco := recommender.NewNeighborhoodBasedRecommender(data, 2)
	recommendations, err := reco.Recommend([]float64{0.0, 0.0, 0.0, 0.0, 1.0, 0.0}, 1, vm.Pearson, false, false, true)
	if err != nil {
		log.Fatalf("Error while recommending: %v", err)
		return
	}

	//fmt.Printf("Recommended items: %v", recommendations)
	fmt.Printf("Recommended items\n")
	for _, recommendation := range recommendations {
		fmt.Printf(recommendation.String())
	}

	recommendations, err = reco.Recommend([]float64{0.0, 0.0, 0.0, 0.0, 1.0, 0.0}, 4, vm.Pearson, true, false, false)
	if err != nil {
		log.Fatalf("Error while recommending: %v", err)
		return
	}

	//fmt.Printf("Recommended items: %v", recommendations)
	fmt.Printf("Recommended items\n")
	for _, recommendation := range recommendations {
		fmt.Printf(recommendation.String())
	}
}
