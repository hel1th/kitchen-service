package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/hel1th/kitchen-service/internal/api/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrdersAPI(t *testing.T) {
	env := SetupE2E(t)
	var userID int64 = 2001
	dishID := "33333333-3333-3333-3333-333333333301" // Пепперони

	// Helper to add item to cart
	addItemToCart := func() {
		payload := map[string]any{"dishId": dishID, "quantity": 1}
		payloadBytes, _ := json.Marshal(payload)
		resp := MakeRequest(t, env.Server, http.MethodPost, "/cart/items", userID, payloadBytes)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}

	t.Run("POST /orders - checkout", func(t *testing.T) {
		// Ensure cart has items
		addItemToCart()

		resp := MakeRequest(t, env.Server, http.MethodPost, "/orders", userID, nil)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var order gen.Order
		err := json.Unmarshal(body, &order)
		require.NoError(t, err)

		assert.Equal(t, gen.Created, *order.Status)
		assert.NotNil(t, order.Id)
		
		// Verify cart is cleared
		respCart := MakeRequest(t, env.Server, http.MethodGet, "/cart", userID, nil)
		cartBody, _ := io.ReadAll(respCart.Body)
		respCart.Body.Close()
		var cart gen.Cart
		json.Unmarshal(cartBody, &cart)
		assert.True(t, cart.Items == nil || len(*cart.Items) == 0)
	})

	t.Run("GET /orders - list orders", func(t *testing.T) {
		resp := MakeRequest(t, env.Server, http.MethodGet, "/orders", userID, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var orders gen.OrderList
		json.Unmarshal(body, &orders)

		require.NotNil(t, orders.Items)
		assert.GreaterOrEqual(t, len(*orders.Items), 1)
	})

	t.Run("PATCH /orders/{orderId}/status - update order status", func(t *testing.T) {
		// Setup new order
		addItemToCart()
		respPost := MakeRequest(t, env.Server, http.MethodPost, "/orders", userID, nil)
		body, _ := io.ReadAll(respPost.Body)
		respPost.Body.Close()
		var order gen.Order
		json.Unmarshal(body, &order)
		orderID := order.Id.String()

		// Update status
		statusPayload := map[string]string{"status": string(gen.Cooking)}
		statusBytes, _ := json.Marshal(statusPayload)
		respPatch := MakeRequest(t, env.Server, http.MethodPatch, "/orders/"+orderID+"/status", 0, statusBytes)
		require.Equal(t, http.StatusOK, respPatch.StatusCode)
		
		patchBody, _ := io.ReadAll(respPatch.Body)
		respPatch.Body.Close()
		var updatedOrder gen.Order
		json.Unmarshal(patchBody, &updatedOrder)

		assert.Equal(t, gen.Cooking, *updatedOrder.Status)
	})
}
