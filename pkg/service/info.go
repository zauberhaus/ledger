package service

import (
	"net/http"
	"sort"

	"github.com/ec-systems/core.ledger.service/pkg/ledger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type InfoService struct {
	chi.Router
	ledger *ledger.Ledger
}

func NewInfoService(ledger *ledger.Ledger) chi.Router {
	router := chi.NewRouter()
	svc := &InfoService{
		Router: router,
		ledger: ledger,
	}

	router.Get("/assets", svc.assets)
	router.Get("/statuses", svc.status)

	return svc
}

// @Summary      Supported Assets
// @Description  List of the assets supported by the ledger
// @Tags         Info
// @Produce      json
// @Success      200  {array}  service.Asset
// @Router       /info/assets [get]
func (i *InfoService) assets(w http.ResponseWriter, r *http.Request) {
	assets := i.ledger.SupportedAssets()
	al := Assets{}

	for k, v := range assets {
		al = append(al, Asset{
			Symbol: k.String(),
			Name:   v,
		})
	}

	sort.Sort(al)

	render.JSON(w, r, al)
}

// @Summary      Supported Statuses
// @Description  List of the statuses supported by the ledger
// @Tags         Info
// @Produce      json
// @Success      200  {array}  service.Statuses
// @Router       /info/statuses [get]
func (i *InfoService) status(w http.ResponseWriter, r *http.Request) {
	statuses := i.ledger.SupportedStatus()
	sl := Statuses{}

	for k, v := range statuses {
		sl = append(sl, Status{int(v), k})
	}

	sort.Sort(sl)

	render.JSON(w, r, sl)
}
