package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with trimmed base URL", func(t *testing.T) {
		client := NewClient("https://example.com/", "test-token")

		assert.Equal(t, "https://example.com", client.BaseURL)
		assert.Equal(t, "test-token", client.APIToken)
		assert.NotNil(t, client.HTTPClient)
	})

	t.Run("keeps base URL without trailing slash", func(t *testing.T) {
		client := NewClient("https://example.com", "test-token")

		assert.Equal(t, "https://example.com", client.BaseURL)
	})
}

func TestNewBasicAuthClient(t *testing.T) {
	t.Run("creates client with trimmed base URL and no API token", func(t *testing.T) {
		client := NewBasicAuthClient("https://example.com/", "test-user", "test-pass")

		assert.Equal(t, "https://example.com", client.BaseURL)
		assert.Equal(t, "test-user", client.Username)
		assert.Equal(t, "test-pass", client.Password)
		assert.Empty(t, client.APIToken)
		assert.NotNil(t, client.HTTPClient)
	})

	t.Run("sends Basic Auth header instead of Bearer", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "test-user", user)
			assert.Equal(t, "test-pass", pass)

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		client := NewBasicAuthClient(server.URL, "test-user", "test-pass")
		_, err := client.Get(context.Background(), "/resources/task", nil)
		require.NoError(t, err)
	})
}

func TestClient_Get(t *testing.T) {
	t.Run("successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "/resources/task", r.URL.Path)
			assert.Equal(t, "test-task", r.URL.Query().Get("taskname"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"sysId": "123",
				"name":  "test-task",
			})
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token")

		query := url.Values{}
		query.Set("taskname", "test-task")

		body, err := client.Get(context.Background(), "/resources/task", query)
		require.NoError(t, err)

		var result map[string]string
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)
		assert.Equal(t, "123", result["sysId"])
		assert.Equal(t, "test-task", result["name"])
	})

	t.Run("GET request without query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/resources/task", r.URL.Path)
			assert.Empty(t, r.URL.RawQuery)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token")
		body, err := client.Get(context.Background(), "/resources/task", nil)

		require.NoError(t, err)
		assert.Contains(t, string(body), "ok")
	})
}

func TestClient_Post(t *testing.T) {
	t.Run("successful POST request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "new-task", body["name"])
			assert.Equal(t, "taskUnix", body["type"])

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Successfully created the task"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token")

		payload := map[string]string{"name": "new-task", "type": "taskUnix"}
		body, err := client.Post(context.Background(), "/resources/task", payload)

		require.NoError(t, err)
		assert.Contains(t, string(body), "Successfully created")
	})
}

func TestClient_Put(t *testing.T) {
	t.Run("successful PUT request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "123", body["sysId"])
			assert.Equal(t, "updated-task", body["name"])

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Successfully updated the task"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token")

		payload := map[string]string{"sysId": "123", "name": "updated-task"}
		body, err := client.Put(context.Background(), "/resources/task", payload)

		require.NoError(t, err)
		assert.Contains(t, string(body), "Successfully updated")
	})
}

func TestClient_Delete(t *testing.T) {
	t.Run("successful DELETE request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "123", r.URL.Query().Get("taskid"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Successfully deleted the task"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token")

		query := url.Values{}
		query.Set("taskid", "123")

		body, err := client.Delete(context.Background(), "/resources/task", query)

		require.NoError(t, err)
		assert.Contains(t, string(body), "Successfully deleted")
	})
}

func TestClient_APIError(t *testing.T) {
	t.Run("returns APIError for 4xx status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Task not found"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token")

		_, err := client.Get(context.Background(), "/resources/task", nil)

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok, "expected APIError type")
		assert.Equal(t, 404, apiErr.StatusCode)
		assert.Equal(t, "Task not found", apiErr.Message)
	})

	t.Run("returns APIError for 5xx status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token")

		_, err := client.Post(context.Background(), "/resources/task", map[string]string{})

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 500, apiErr.StatusCode)
	})
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "not found"}
	assert.Equal(t, "API error (status 404): not found", err.Error())
}
