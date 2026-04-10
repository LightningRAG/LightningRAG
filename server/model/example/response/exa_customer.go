package response

import "github.com/LightningRAG/LightningRAG/server/model/example"

type ExaCustomerResponse struct {
	Customer example.ExaCustomer `json:"customer"`
}
