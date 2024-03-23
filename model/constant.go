package model

const (
	ECodeInternal      = "50001_" // for error when internal server error
	ECodeBadRequest    = "40001_" // for error when failed decode request
	ECodeNotFound      = "40002_" // for error when resource not found
	ECodeValidateFail  = "40003_" // for error when request is failed validation
	ECodeMethodFail    = "40004_" // for error when request is method not allowed
	ECodeDataExists    = "40005_" // for error when data is duplicate or exists
	ECodeAuthorization = "40006_" // for error when request need authorization
	ECodeForbidden     = "40007_" // for error when request cannot be access because some reason
)
