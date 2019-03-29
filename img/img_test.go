package img

import (
	"context"
	"testing"

	"github.com/yansal/img/storage/backends/local"
)

func Test(t *testing.T) {
	m := &Processor{storage: &local.Storage{}}
	_, err := m.process(context.Background(), Payload{
		URL:    "https://upload.wikimedia.org/wikipedia/commons/5/57/Villach_Sankt_Leonhard_Pfarrkirche_hl._Leonhard_Altarwandgem%C3%A4lde_Gnadenstuhl_24092018_4768.jpg",
		Width:  230,
		Height: 126,
	})
	if err != nil {
		t.Error(err)
	}
}
