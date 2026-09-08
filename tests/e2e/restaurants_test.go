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

func TestRestaurantsAPI(t *testing.T) {
	env := SetupE2E(t)

	// Seeded restaurant ID
	restID := "11111111-1111-1111-1111-111111111111"

	t.Run("GET /restaurants - list restaurants", func(t *testing.T) {
		resp := MakeRequest(t, env.Server, http.MethodGet, "/restaurants", 0, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var listResp gen.RestaurantList
		err = json.Unmarshal(body, &listResp)
		require.NoError(t, err)

		require.NotNil(t, listResp.Items)
		assert.GreaterOrEqual(t, len(*listResp.Items), 1)
	})

	t.Run("GET /restaurants/{id} - get single restaurant", func(t *testing.T) {
		resp := MakeRequest(t, env.Server, http.MethodGet, "/restaurants/"+restID, 0, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var restResp gen.Restaurant
		err = json.Unmarshal(body, &restResp)
		require.NoError(t, err)

		assert.Equal(t, restID, restResp.Id.String())
		assert.Equal(t, gen.Open, *restResp.Status)
	})

	t.Run("PATCH /restaurants/{id}/status - close restaurant", func(t *testing.T) {
		payload := map[string]string{"status": string(gen.Closed)}
		payloadBytes, _ := json.Marshal(payload)

		resp := MakeRequest(t, env.Server, http.MethodPatch, "/restaurants/"+restID+"/status", 0, payloadBytes)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var restResp gen.Restaurant
		err = json.Unmarshal(body, &restResp)
		require.NoError(t, err)

		assert.Equal(t, gen.Closed, *restResp.Status)
		// Revert status to open for other tests
		payload = map[string]string{"status": string(gen.Open)}
		payloadBytes, _ = json.Marshal(payload)
		resp2 := MakeRequest(t, env.Server, http.MethodPatch, "/restaurants/"+restID+"/status", 0, payloadBytes)
		require.Equal(t, http.StatusOK, resp2.StatusCode)
		resp2.Body.Close()
	})
}
