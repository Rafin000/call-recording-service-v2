package domain

// Params struct to represent the parameters including `i_customer`
// type Params struct {
// 	ICustomer string `json:"i_customer"`
// }

// XDRRequest struct to wrap `auth_info` and `params` fields
type XDRRequest struct {
	ICustomer string `json:"i_customer"`
}

// XDRDumpsRequest struct for request parameters related to XDR dumps
type XDRDumpsRequest struct {
	Page      int    `json:"page" binding:"required"`
	PageSize  int    `json:"page_size" binding:"required"`
	FromDate  string `json:"from_date" binding:"required"`
	ToDate    string `json:"to_date" binding:"required"`
	ICustomer string `json:"i_customer" binding:"required"`
}
