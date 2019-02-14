package main

import (
	"context"
	"testing"

	"github.com/yansal/img/storage/backends/local"
)

func Test(t *testing.T) {
	m := &manager{cache: true, storage: &local.Storage{}}
	_, err := m.process(context.Background(), payload{
		url:    "https://upload.wikimedia.org/wikipedia/commons/5/57/Villach_Sankt_Leonhard_Pfarrkirche_hl._Leonhard_Altarwandgem%C3%A4lde_Gnadenstuhl_24092018_4768.jpg",
		width:  230,
		height: 126,
		cache:  true,
	})
	if err != nil {
		t.Error(err)
	}
}
