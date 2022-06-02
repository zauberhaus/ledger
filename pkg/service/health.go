package service

import (
	"net/http"

	"github.com/ec-systems/core.ledger.service/pkg/ledger"
	"github.com/go-chi/render"
)

type HealthService struct {
	ledger *ledger.Ledger
}

func NewHealthService(ledger *ledger.Ledger) map[string]http.HandlerFunc {
	svc := &HealthService{
		ledger: ledger,
	}

	return map[string]http.HandlerFunc{
		"/health": svc.health,
	}
}

// @Summary      Health
// @Description  Show health status
// @Tags         Health
// @Produce      json
// @Success      200  string
// @Failure      500
// @Router       /health [get]
func (h *HealthService) health(w http.ResponseWriter, r *http.Request) {
	_, err := h.ledger.Health(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.PlainText(w, r, "ok")
}
