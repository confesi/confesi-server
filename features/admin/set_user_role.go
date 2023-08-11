package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Attempts to execute the cron job once, without retries.
func (h *handler) handleSetUserRole(c *gin.Context) {
	// Validate request
	var req validation.UpdateUserRole
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// Obtain User ID
	uid := req.UserID

	// Check if user exists
	user, err := h.fb.AuthClient.GetUser(c, uid)
	if err != nil {
		response.New(http.StatusNotFound).Err("user not found").Send(c)
		return
	}

	// Obtain Action User ID
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// Obtain Roles
	roles := user.CustomClaims["roles"].([]interface{})

	// Create Role Logs
	roleLogs := db.RoleAssignmentLog{
		ActionType:     req.Action,
		ActionUserID:   token.UID,
		AffectedUserID: uid,
		OldRoles:       interface_to_string(roles),
	}

	if req.Action == "remove" {
		// Remove roles
		temp := make([]interface{}, 0)

		// Iterate through user current roles
		for _, item := range roles {
			found := false
			// Check if item is in req.Roles (removal roles)
			for _, rItem := range req.Roles {
				if item == rItem {
					found = true
					break
				}
			}
			// If not found, add to temp
			if !found {
				temp = append(temp, item)
			}
		}

		roles = temp
	} else if req.Action == "add" {
		// Add role (Append)
		for _, role := range req.Roles {
			roles = append(roles, role)
		}

	} else {
		// Set Roles (Overwrite)
		temp := make([]interface{}, len(req.Roles))
		for i := range req.Roles {
			temp[i] = req.Roles[i]
		}
		roles = temp
	}

	// Remove duplicates
	roles = unique_list(roles)

	// Update Role Logs
	roleLogs.NewRoles = interface_to_string(roles)

	err = h.db.Create(&roleLogs).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// Set Custom Claims (Update User Roles)
	err = h.fb.AuthClient.SetCustomUserClaims(c, uid, map[string]interface{}{
		"sync":  true,
		"roles": roles,
	})
	if err != nil {
		response.New(http.StatusNotFound).Err("error updating user").Send(c)
		return
	}

	// If everything went well return success
	response.New(http.StatusOK).Send(c)
}

// Create unique lists (no dupes)
func unique_list(interface_list []interface{}) []interface{} {
	stringSlice := make([]string, len(interface_list))
	for i, v := range interface_list {
		stringSlice[i] = v.(string)
	}
	keys := make(map[string]bool)
	list := make([]interface{}, 0)
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func interface_to_string(interface_list []interface{}) string {
	output := "{"
	for i, item := range interface_list {
		if i == len(interface_list)-1 {
			output += item.(string)
		} else {
			output += item.(string) + ","
		}

	}
	output += "}"
	return output
}
