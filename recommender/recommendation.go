package recommender

import (
	"fmt"

	vm "github.com/juandes/knn-recommender-system/vectormath"
)

type Recommendation interface {
	String() string
	Distance() float64
	Items() []float64
}

type MultipleRecommendation struct {
	index    int
	items    []float64
	d        float64
	distance vm.Distance
}

type SingleRecommendation struct {
	item     []float64
	distance vm.Distance
}

type SerendipitousRecommendation struct {
	item     []float64
	distance vm.Distance
}

func (r MultipleRecommendation) String() string {
	return fmt.Sprintf("Items: %v\nIndex: %d\nDistance (%v): %f\n", r.items, r.index, r.distance, r.d)
}

func (r MultipleRecommendation) Distance() float64 {
	return r.d
}

func (r MultipleRecommendation) Items() []float64 {
	return r.items
}

func (r SingleRecommendation) String() string {
	return fmt.Sprintf("Items: %v\nDistance used: %v\n", r.item, r.distance)
}

func (r SingleRecommendation) Distance() float64 {
	return 0.0
}

func (r SingleRecommendation) Items() []float64 {
	return r.item
}

func (r SerendipitousRecommendation) String() string {
	return fmt.Sprintf("Serendipitous recommendation items: %v\nDistance used: %v\n", r.item, r.distance)
}

func (r SerendipitousRecommendation) Distance() float64 {
	return 0.0
}

func (r SerendipitousRecommendation) Items() []float64 {
	return r.item
}
