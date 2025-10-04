package response

import (
	"github.com/si/internal/types"

)


func ErrorResponse(message string, details ...interface{}) types.APIResponse {
    return types.APIResponse{
        Message: message,
        Error: &types.ErrorResponse{
            Details: details,
        },
    }
}
