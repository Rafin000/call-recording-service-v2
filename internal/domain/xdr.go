package domain

type XDRRequest struct {
	ICustomer string `json:"i_customer" binding:"required"`
}

type XDRDumpsRequest struct {
	Page      int    `json:"page" binding:"required"`
	PageSize  int    `json:"page_size" binding:"required"`
	FromDate  string `json:"from_date" binding:"required"`
	ToDate    string `json:"to_date" binding:"required"`
	ICustomer string `json:"i_customer" binding:"required"`
}
