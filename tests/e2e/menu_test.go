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

func TestMenuAPI(t *testing.T) {
	env := SetupE2E(t)
	restID := "11111111-1111-1111-1111-111111111111"

	t.Run("GET /restaurants/{id}/menu - get menu", func(t *testing.T) {
		resp := MakeRequest(t, env.Server, http.MethodGet, "/restaurants/"+restID+"/menu", 0, nil)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("GET /menu returned %d: %s", resp.StatusCode, string(body))
		}
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var menu []gen.CategoryWithDishes
		err = json.Unmarshal(body, &menu)
		require.NoError(t, err)

		assert.NotEmpty(t, menu)
		if len(menu) > 0 {
			assert.NotEmpty(t, menu[0].Dishes)
		}
	})

	t.Run("POST /restaurants/{id}/categories - create category", func(t *testing.T) {
		payload := map[string]any{"name": "Desserts", "sortOrder": 3}
		payloadBytes, _ := json.Marshal(payload)

		resp := MakeRequest(t, env.Server, http.MethodPost, "/restaurants/"+restID+"/categories", 0, payloadBytes)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Logf("POST /categories returned %d: %s", resp.StatusCode, string(body))
		}
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var category gen.Category
		err := json.Unmarshal(body, &category)
		require.NoError(t, err)

		assert.Equal(t, "Desserts", *category.Name)

		// DELETE the category
		respDel := MakeRequest(t, env.Server, http.MethodDelete, "/restaurants/"+restID+"/categories/"+category.Id.String(), 0, nil)
		require.Equal(t, http.StatusNoContent, respDel.StatusCode)
		respDel.Body.Close()
	})

	t.Run("POST /restaurants/{id}/dishes - create and manipulate dish", func(t *testing.T) {
		// 1. Create a category for the dish
		catPayload := map[string]any{"name": "New Category", "sortOrder": 4}
		catBytes, _ := json.Marshal(catPayload)
		catResp := MakeRequest(t, env.Server, http.MethodPost, "/restaurants/"+restID+"/categories", 0, catBytes)
		catBody, _ := io.ReadAll(catResp.Body)
		catResp.Body.Close()
		var category gen.Category
		json.Unmarshal(catBody, &category)

		// 2. Create dish
		dishPayload := gen.DishInput{
			CategoryId: *category.Id,
			Name:       "Test Dish",
			Price:      100.50,
		}
		dishBytes, _ := json.Marshal(dishPayload)
		resp := MakeRequest(t, env.Server, http.MethodPost, "/restaurants/"+restID+"/dishes", 0, dishBytes)
		dishBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Logf("POST /dishes returned %d: %s", resp.StatusCode, string(dishBody))
		}
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var dish gen.Dish
		json.Unmarshal(dishBody, &dish)

		assert.Equal(t, "Test Dish", *dish.Name)
		assert.True(t, *dish.Available)

		// 3. Patch dish availability
		availPayload := map[string]bool{"available": false}
		availBytes, _ := json.Marshal(availPayload)
		respAvail := MakeRequest(t, env.Server, http.MethodPatch, "/restaurants/"+restID+"/dishes/"+dish.Id.String()+"/availability", 0, availBytes)
		require.Equal(t, http.StatusOK, respAvail.StatusCode)
		respAvail.Body.Close()

		// 4. Delete dish (soft delete)
		respDel := MakeRequest(t, env.Server, http.MethodDelete, "/restaurants/"+restID+"/dishes/"+dish.Id.String(), 0, nil)
		require.Equal(t, http.StatusNoContent, respDel.StatusCode)
		respDel.Body.Close()

		// 5. Restore dish
		respRestore := MakeRequest(t, env.Server, http.MethodPost, "/restaurants/"+restID+"/dishes/"+dish.Id.String()+"/restore", 0, nil)
		require.Equal(t, http.StatusOK, respRestore.StatusCode)
		respRestore.Body.Close()
	})
}
