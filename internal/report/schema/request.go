package schema

type RequestReportList struct {
	// pagination
	Page     int `form:"page" validate:"required,min=1"`
	PageSize int `form:"page_size" validate:"required,min=1,max=100"`
}
