package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func BadRequestError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, message)
}

func NotFoundError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message)
}

func InternalServerError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, message)
}

func UnauthorizedError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message)
}

func ForbiddenError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message)
}

func ConflictError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusConflict, message)
}

func OKResponse(c *gin.Context, data interface{}) {
	SuccessResponse(c, http.StatusOK, data)
}

func CreatedResponse(c *gin.Context, data interface{}) {
	SuccessResponse(c, http.StatusCreated, data)
}

func NoContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func TooManyRequestsError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusTooManyRequests, message)
}

func PaymentRequiredError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusPaymentRequired, message)
}

// QuotaExceededResponse returns a 402 with structured quota details so the
// frontend can surface upgrade prompts per resource.
func QuotaExceededResponse(c *gin.Context, resource string, limit, usage int) {
	c.JSON(http.StatusPaymentRequired, gin.H{
		"error":    "quota_exceeded",
		"resource": resource,
		"limit":    limit,
		"usage":    usage,
	})
}

func ServiceUnavailableError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusServiceUnavailable, message)
}
