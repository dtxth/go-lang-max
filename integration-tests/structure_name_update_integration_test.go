package integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UpdateNameRequest represents request to update name
type UpdateNameRequest struct {
	Name string `json:"name"`
}

// StructureNode represents a node in the hierarchical structure
type StructureNode struct {
	Type      string          `json:"type"`
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Children  []*StructureNode `json:"children,omitempty"`
	Chat      interface{}     `json:"chat,omitempty"`
	Course    *int            `json:"course,omitempty"`
	GroupNum  *string         `json:"group_num,omitempty"`
	ChatCount *int            `json:"chat_count"`
}

func TestStructureNameUpdate_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseURL := "http://localhost:8083"
	
	// Wait for service to be ready
	waitForService(t, baseURL+"/universities", 30*time.Second)

	t.Run("UpdateUniversityName", func(t *testing.T) {
		// Get initial structure to find university ID
		resp, err := http.Get(baseURL + "/universities/1/structure")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		if resp.StatusCode == 404 {
			t.Skip("University with ID 1 not found, skipping test")
		}
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var initialStructure StructureNode
		err = json.NewDecoder(resp.Body).Decode(&initialStructure)
		require.NoError(t, err)

		originalName := initialStructure.Name
		newName := fmt.Sprintf("Updated University Name - %d", time.Now().Unix())

		// Update university name
		updateReq := UpdateNameRequest{Name: newName}
		jsonData, err := json.Marshal(updateReq)
		require.NoError(t, err)

		resp, err = http.Post(baseURL+"/universities/1/name", "application/json", bytes.NewBuffer(jsonData))
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		}
		
		// Use PUT method instead of POST
		req, err := http.NewRequest("PUT", baseURL+"/universities/1/name", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the name was updated
		resp, err = http.Get(baseURL + "/universities/1/structure")
		require.NoError(t, err)
		defer resp.Body.Close()

		var updatedStructure StructureNode
		err = json.NewDecoder(resp.Body).Decode(&updatedStructure)
		require.NoError(t, err)

		assert.Equal(t, newName, updatedStructure.Name)
		assert.NotEqual(t, originalName, updatedStructure.Name)

		// Restore original name
		restoreReq := UpdateNameRequest{Name: originalName}
		jsonData, err = json.Marshal(restoreReq)
		require.NoError(t, err)

		req, err = http.NewRequest("PUT", baseURL+"/universities/1/name", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err = client.Do(req)
		if err == nil && resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	})

	t.Run("UpdateGroupName", func(t *testing.T) {
		// Get structure to find a group
		resp, err := http.Get(baseURL + "/universities/1/structure")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		if resp.StatusCode == 404 {
			t.Skip("University with ID 1 not found, skipping test")
		}

		var structure StructureNode
		err = json.NewDecoder(resp.Body).Decode(&structure)
		require.NoError(t, err)

		// Find first group in the structure
		var groupID int64
		var originalGroupName string
		found := false

		for _, branch := range structure.Children {
			if found {
				break
			}
			for _, faculty := range branch.Children {
				if found {
					break
				}
				for _, group := range faculty.Children {
					if group.Type == "group" {
						groupID = group.ID
						if group.GroupNum != nil {
							originalGroupName = *group.GroupNum
						} else {
							originalGroupName = group.Name
						}
						found = true
						break
					}
				}
			}
		}

		if !found {
			t.Skip("No groups found in structure, skipping test")
		}

		newGroupName := fmt.Sprintf("Updated-Group-%d", time.Now().Unix())

		// Update group name
		updateReq := UpdateNameRequest{Name: newGroupName}
		jsonData, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/groups/%d/name", baseURL, groupID), bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the group name was updated
		resp, err = http.Get(baseURL + "/universities/1/structure")
		require.NoError(t, err)
		defer resp.Body.Close()

		var updatedStructure StructureNode
		err = json.NewDecoder(resp.Body).Decode(&updatedStructure)
		require.NoError(t, err)

		// Find the updated group
		groupFound := false
		for _, branch := range updatedStructure.Children {
			if groupFound {
				break
			}
			for _, faculty := range branch.Children {
				if groupFound {
					break
				}
				for _, group := range faculty.Children {
					if group.Type == "group" && group.ID == groupID {
						if group.GroupNum != nil {
							assert.Equal(t, newGroupName, *group.GroupNum)
						} else {
							assert.Equal(t, newGroupName, group.Name)
						}
						groupFound = true
						break
					}
				}
			}
		}

		assert.True(t, groupFound, "Updated group not found in structure")

		// Restore original name
		restoreReq := UpdateNameRequest{Name: originalGroupName}
		jsonData, err = json.Marshal(restoreReq)
		require.NoError(t, err)

		req, err = http.NewRequest("PUT", fmt.Sprintf("%s/groups/%d/name", baseURL, groupID), bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err = client.Do(req)
		if err == nil && resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	})

	t.Run("UpdateNonExistentUniversity", func(t *testing.T) {
		updateReq := UpdateNameRequest{Name: "Non-existent University"}
		jsonData, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", baseURL+"/universities/99999/name", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("UpdateWithEmptyName", func(t *testing.T) {
		updateReq := UpdateNameRequest{Name: "   "}
		jsonData, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", baseURL+"/universities/1/name", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UpdateWithInvalidID", func(t *testing.T) {
		updateReq := UpdateNameRequest{Name: "Valid Name"}
		jsonData, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", baseURL+"/universities/invalid/name", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func waitForService(t *testing.T, url string, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode < 500 {
			if resp.Body != nil {
				resp.Body.Close()
			}
			return
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Service not ready after %v", timeout)
}