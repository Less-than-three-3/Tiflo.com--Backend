package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strings"
	"tiflo/model"
)

func (h *Handler) AuthCheck() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		jwtStr, err := gCtx.Cookie("AccessToken")
		if err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		if !strings.HasPrefix(jwtStr, model.JwtPrefix) {
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		if len(jwtStr) != 0 {
			jwtStr = jwtStr[len(model.JwtPrefix):]
		}
		err = h.redisClient.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr)
		if err == nil {
			gCtx.AbortWithStatus(http.StatusForbidden)

			return
		}
		if !errors.Is(err, redis.Nil) {
			gCtx.AbortWithError(http.StatusInternalServerError, err)

			return
		}

		userId, err := h.tokenManager.Parse(jwtStr)

		if err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}

		gCtx.Set(model.UserCtx, userId)
		gCtx.Next()
	}
}
