package recommender

import (
	"fmt"
	"github.com/juandes/knn-recommender-system/vectormath"
	"math"
	"math/rand"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	vm "github.com/juandes/knn-recommender-system/vectormath"
)

type NeighborhoodBasedRecommender struct {
	data        [][]float64
	neighbors   int
	numberItems int
}

type Slice struct {
	sort.Interface
	idx []int
}

// NewNeighborhoodBasedRecommender creates a new NeighborhoodBasedRecommender object
func NewNeighborhoodBasedRecommender(data [][]float64, k int) *NeighborhoodBasedRecommender {
	if len(data) == 0 {
		log.Fatalf("Dataset is empty")
	}

	return &NeighborhoodBasedRecommender{
		data:        data,
		neighbors:   k,
		numberItems: len(data[0]),
	}
}

// Recommend recommends the n number of items that are closer to a given vector using a given distance measure
func (nbr *NeighborhoodBasedRecommender) Recommend(items []float64, numItemsToRecommend int, distanceMeasure vm.Distance, intercept, shuffle, serendipitous bool) ([]Recommendation, error) {
	// TODO (Juan): If vector is a zero vector, it should return
	recommendations, err := nbr.findKNearestNeighbors(items, distanceMeasure, intercept, shuffle, serendipitous)
	if err != nil {
		return nil, fmt.Errorf("Error encountered while finding K nearest neighbors: %v", err)
	}

	return recommendations, nil
}

func (nbr *NeighborhoodBasedRecommender) findKNearestNeighbors(items []float64, distanceMeasure vm.Distance, intercept, shuffle, serendipitous bool) ([]Recommendation, error) {
	var (
		d                 float64
		err               error
		distancesFromUser []Recommendation
		recommendations   []Recommendation
		order             []int
	)

	// order is an array where the values are
	// 1 ...n where n is the number of rows
	// in the training dataset.
	// It is the equivalent of Python's range(len(nbr.data))
	order = make([]int, len(nbr.data))
	for i := range order {
		order[i] = i
	}

	// the point of shuffling the order in which
	// the distances will be calculated
	// is to avoid having always the same
	// predictions in case all the n results
	// to return have the same distance
	if shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	}

	for i, val := range order {
		user := nbr.data[val]
		if len(user) != nbr.numberItems {
			return nil, fmt.Errorf("Incorrect number of items in vector")
		}

		switch distanceMeasure {
		case vm.Euclidean:
			d, err = vm.EuclideanDistance(items, user)
		case vm.Cosine:
			d, err = vm.CosineSimilarity(items, user)
		case vm.Manhattan:
			d, err = vm.ManhattanDistance(items, user)
		case vm.Pearson:
			d, err = vm.PearsonCorrelation(items, user)
			// The Pearson correlation coefficient lies between -1 and 1,
			// however I want the score to be from 0 and 1, where
			// 0 represents perfect correlation regardless of whether
			// it is a positive or negative one
			d = 1 - math.Abs(d)
		default:
			return nil, fmt.Errorf("Invalid distance measure: %v", distanceMeasure)
		}

		if err != nil {
			return nil, fmt.Errorf("Error calculating distance: %v", err)
		}

		distancesFromUser = append(distancesFromUser, MultipleRecommendation{
			index:    i,
			items:    user,
			d:        d,
			distance: distanceMeasure,
		})
	}

	// sort the recommendations (ascending order) by distance from the given vector
	sort.Slice(distancesFromUser, func(i, j int) bool {
		return distancesFromUser[i].Distance() < distancesFromUser[j].Distance()
	})
	recommendations = distancesFromUser[:nbr.neighbors]

	// The idea here is the following:
	// 1. Get the n:n*2 neighbors
	// 2. Build a map where the keys are the champions found on those neighbors
	//    and value is the count of them.
	// 3. Sort the map by its value
	// 4. Use the 5 champions with the highest count as a recommendation
	if serendipitous {
		sereOptions := make([]float64, len(recommendations[0].Items()))
		for _, reco := range distancesFromUser[nbr.neighbors:int(math.Min(float64(nbr.neighbors*2), float64(len(nbr.data))))] {
			//log.Infof("i: %v", int(math.Min(float64(nbr.neighbors*2), float64(len(nbr.data)))))
			for j, item := range reco.Items() {
				log.Infof("j: %v", j)
				sereOptions[j] += item
			}
		}
		log.Printf("sereOptions: %v", sereOptions)
		s := NewFloat64Slice(sereOptions...)
		sort.Sort(sort.Reverse(s))
		log.Printf("sereOptions (sorted): %v", s.idx)
		serendipitousRecommendation := make([]float64, len(recommendations[0].Items()))
		// 5 because we are interested in the top 5 champions
		for _, val := range s.idx[0:5] {
			serendipitousRecommendation[val] = 1
		}

		recommendations = append(recommendations, SerendipitousRecommendation{
			item:     serendipitousRecommendation,
			distance: distanceMeasure,
		})

	} else if intercept { // serendipitous and intercept modes recommendations are mutually exclusive
		intercepts := recommendations[0].Items()
		for _, val := range recommendations[1:len(recommendations)] {
			intercepts, err = vectormath.SetIntercept(intercepts, val.Items())
			if err != nil {
				return nil, fmt.Errorf("Error calculating set intercept: %v", err)
			}
		}

		recommendations = []Recommendation{
			&SimpleRecommendation{
				item:     intercepts,
				distance: distanceMeasure,
			},
		}

	}

	return recommendations, nil
}

func NewFloat64Slice(n ...float64) *Slice { return NewSlice(sort.Float64Slice(n)) }

func NewSlice(n sort.Interface) *Slice {
	s := &Slice{
		Interface: n,
		idx:       make([]int, n.Len()),
	}

	for i := range s.idx {
		s.idx[i] = i
	}
	return s
}

func (s Slice) Swap(i, j int) {
	s.Interface.Swap(i, j)
	s.idx[i], s.idx[j] = s.idx[j], s.idx[i]
}
