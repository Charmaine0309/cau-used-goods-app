package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"cau-used-goods-app/backend/pkg/response"
)

func ReadableAccount(db *sql.DB) gin.HandlerFunc {
	return requireAccount(db, false)
}

func NormalAccount(db *sql.DB) gin.HandlerFunc {
	return requireAccount(db, true)
}

func requireAccount(db *sql.DB, normalOnly bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或登录状态已失效")
			c.Abort()
			return
		}

		var accountStatus string
		var isDeleted bool
		err := db.QueryRowContext(
			c.Request.Context(),
			`SELECT account_status, is_deleted FROM users WHERE id = ? LIMIT 1`,
			userID,
		).Scan(&accountStatus, &isDeleted)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户不存在")
			} else {
				response.Error(c, http.StatusInternalServerError, response.CodeInternal, "校验账号状态失败")
			}
			c.Abort()
			return
		}

		if isDeleted || accountStatus == "CANCELED" {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "账号已注销")
			c.Abort()
			return
		}
		if normalOnly && accountStatus != "NORMAL" {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "当前账号状态不可操作")
			c.Abort()
			return
		}
		c.Next()
	}
}
