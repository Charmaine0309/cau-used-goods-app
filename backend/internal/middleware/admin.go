package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"cau-used-goods-app/backend/pkg/response"
)

func Admin(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
			c.Abort()
			return
		}

		var role, accountStatus string
		err := db.QueryRowContext(
			c.Request.Context(),
			`SELECT role, account_status FROM users WHERE id = ? AND is_deleted = 0 LIMIT 1`,
			userID,
		).Scan(&role, &accountStatus)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.Error(c, http.StatusForbidden, response.CodeForbidden, "admin permission required")
			} else {
				response.Error(c, http.StatusInternalServerError, response.CodeInternal, "check admin permission failed")
			}
			c.Abort()
			return
		}

		if role != "ADMIN" {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "admin permission required")
			c.Abort()
			return
		}
		if accountStatus != "NORMAL" {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "当前管理员账号状态不可操作")
			c.Abort()
			return
		}
		c.Next()
	}
}
