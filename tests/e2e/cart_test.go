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

func TestCartAPI(t *testing.T) {
	env := SetupE2E(t)
	// User ID for cart testing
	var userID int64 = 1001
	// Seeded dish ID
	dishID := "33333333-3333-3333-3333-333333333301"

	t.Run("GET /cart - empty cart", func(t *testing.T) {
		resp := MakeRequest(t, env.Server, http.MethodGet, "/cart", userID, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var cart gen.Cart
		json.Unmarshal(body, &cart)

		assert.True(t, cart.Items == nil || len(*cart.Items) == 0)
	})

	var addedItemID string

	t.Run("POST /cart/items - add dish to cart", func(t *testing.T) {
		payload := map[string]any{"dishId": dishID, "quantity": 2}
		payloadBytes, _ := json.Marshal(payload)

		resp := MakeRequest(t, env.Server, http.MethodPost, "/cart/items", userID, payloadBytes)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var cart gen.Cart
		err := json.Unmarshal(body, &cart)
		require.NoError(t, err)

		require.NotNil(t, cart.Items)
		assert.Len(t, *cart.Items, 1)
		assert.Equal(t, 2, *(*cart.Items)[0].Quantity)
		
		addedItemID = (*cart.Items)[0].Id.String()
	})

	t.Run("PATCH /cart/items/{itemId} - update quantity", func(t *testing.T) {
		payload := map[string]any{"quantity": 5}
		payloadBytes, _ := json.Marshal(payload)

		resp := MakeRequest(t, env.Server, http.MethodPatch, "/cart/items/"+addedItemID, userID, payloadBytes)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var cart gen.Cart
		json.Unmarshal(body, &cart)

		require.NotNil(t, cart.Items)
		assert.Equal(t, 5, *(*cart.Items)[0].Quantity)
	})

	t.Run("DELETE /cart/items/{itemId} - remove item", func(t *testing.T) {
		resp := MakeRequest(t, env.Server, http.MethodDelete, "/cart/items/"+addedItemID, userID, nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

		// Verify empty
		respGet := MakeRequest(t, env.Server, http.MethodGet, "/cart", userID, nil)
		body, _ := io.ReadAll(respGet.Body)
		respGet.Body.Close()
		var cart gen.Cart
		json.Unmarshal(body, &cart)
		assert.True(t, cart.Items == nil || len(*cart.Items) == 0)
	})
}
