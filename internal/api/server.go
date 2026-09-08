package api

import (
	menu_api "github.com/hel1th/kitchen-service/internal/module/menu/api"
	order_api "github.com/hel1th/kitchen-service/internal/module/order/api"
	rest_api "github.com/hel1th/kitchen-service/internal/module/restaurant/api"
)

// Server implements gen.ServerInterface by composing the module handlers via embedding.
type Server struct {
	*menu_api.MenuHandler
	*order_api.OrderHandler
	*rest_api.RestaurantHandler
}

// NewServer creates a new API server composition root.
func NewServer(
	menuHandler *menu_api.MenuHandler,
	orderHandler *order_api.OrderHandler,
	restHandler *rest_api.RestaurantHandler,
) *Server {
	return &Server{
		MenuHandler:       menuHandler,
		OrderHandler:      orderHandler,
		RestaurantHandler: restHandler,
	}
}
