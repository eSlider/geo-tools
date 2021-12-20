package mbtiles

import (
	"testing"
)

func TestSearchingByWord(t *testing.T) {
	manager, err := NewManager("../../data/canary-islands-latest.mbtiles")
	if err != nil {
		t.Error(err)
	}

	searchQuery := "santa"
	features, err := manager.Search(searchQuery, 5)
	if err != nil {
		t.Error(err)
	}

	if len(features) < 1 {
		t.Errorf("nothing found")
	}
}
