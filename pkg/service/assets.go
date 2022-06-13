package service

import (
	"fmt"
	"net/http"

	"github.com/ec-systems/core.ledger.server/pkg/ledger"
	"github.com/ec-systems/core.ledger.server/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type AssetsService struct {
	chi.Router
	ledger *ledger.Ledger
}

func NewAssetsService(ledger *ledger.Ledger) chi.Router {
	router := chi.NewRouter()
	svc := &AssetsService{
		Router: router,
		ledger: ledger,
	}

	// list assets
	router.Get("/", svc.assets)
	// show balance of an asset
	router.Get("/{asset}", svc.balance)

	return svc
}

// @Summary      Show list of Assets
// @Description  Show alle assets with a transaction
// @Tags         Assets
// @Produce      json
// @Success      200  {array}  service.Asset
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /assets/ [get]
func (a *AssetsService) assets(w http.ResponseWriter, r *http.Request) {
	assets, err := a.ledger.Assets(r.Context())
	if isError(w, err) {
		return
	}

	render.JSON(w, r, assets)
}

// @Summary      Asset Balance
// @Description  Show the balance of an Asset
// @Tags         Assets
// @Produce      json
// @Param        asset   	path      	string  true  	"Asset Symbol"
// @Success      200  {object}  service.AssetBalance
// @Failure      400
// @Failure      404
// @Failure      406
// @Failure      500
// @Router       /assets/{asset} [get]
func (a *AssetsService) balance(w http.ResponseWriter, r *http.Request) {
	asset, err := a.asset(w, r)
	if isError(w, err) {
		return
	}

	balance, err := a.ledger.AssetBalance(r.Context(), asset)
	if isError(w, err) {
		return
	}

	result := AssetBalance{}
	for k, v := range balance {
		result = AssetBalance{
			Asset: k.String(),
			Sum:   v,
		}

		render.JSON(w, r, result)
		return
	}

	http.Error(w, fmt.Sprintf("asset '%v' not found", asset), http.StatusNotFound)
}

func (t *AssetsService) asset(w http.ResponseWriter, r *http.Request) (types.Asset, error) {
	assetID := chi.URLParam(r, "asset")

	if assetID == "" {
		return types.AllAssets, nil
	}

	asset, err := t.ledger.SupportedAssets().Parse(assetID)
	if err == nil {
		return asset, nil
	}

	return types.AllAssets, ledger.NewError(http.StatusBadRequest, err.Error())
}
