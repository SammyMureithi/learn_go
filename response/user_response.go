package response

import "my_store_app/models"

type Response struct {
	OK     bool        `json:"ok"`
	Status string      `json:"status"`
	Message string     `json:"message"`
	User   models.User `json:"user"`
}
