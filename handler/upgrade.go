package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Upgrade handles websocket upgrading
func (h *Handler) Upgrade(c echo.Context) error {
	conn, err := h.conf.Upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		log.Error().Err(err).Msg("handler: failed to upgrade http to websocket")

		return c.String(http.StatusInternalServerError, "unexpected error")
	}

	h.conf.Hub.Register(h.conf.ClientFactory.New(conn))

	log.Debug().Msg("handler: sucessfully upgrade connection")

	return nil
}
