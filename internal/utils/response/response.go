package response

import (
	"github.com/omniful/go_commons/http"
	"github.com/si/internal/types"

)


func ErrorResponse(code int, message string, details ...interface{}) types.APIResponse {
    return types.APIResponse{
        Headers: http.ResponseParams{
			StatusCode: code,
		},
        Error: &types.ErrorResponse{
            Message: message,
            Details: details,
        },
    }
}
