package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"tiflo/model"

	"github.com/gin-gonic/gin"
)

// SignIn godoc
// @Summary      User sign-in
// @Description  Authenticates a user and generates an access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        user  body  model.UserLogin  true  "User information"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/signIn [post]
func (h *Handler) SignIn(context *gin.Context) {
	var user model.UserLogin
	var err error

	if err = context.BindJSON(&user); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "неверный формат данных")
		return
	}

	if user.Password, err = h.hasher.Hash(user.Password); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "неверный формат пароля"})
		return
	}

	userInfo, err := h.repo.GetUser(context.Request.Context(), user)
	if err != nil {
		if errors.Is(err, model.NotFound) {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь с таким логином не найден"})
			return
		}

		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ошибка авторизации"})
		return
	}

	token, err := h.tokenManager.NewJWT(userInfo.UserId)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ошибка при формировании токена"})
		return
	}

	context.SetCookie("AccessToken", "Bearer "+token, 0, "/", "localhost", false, true)
	context.JSON(http.StatusOK, gin.H{"message": "клиент успешно авторизован", "login": user.Login, "userId": userInfo.UserId})
}

// SignUp godoc
// @Summary      Sign up a new user
// @Description  Creates a new user account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        user  body  model.UserLogin  true  "User login and password"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      409  {object}  error
// @Failure      500  {object}  error
// @Router       /api/signUp [post]
func (h *Handler) SignUp(context *gin.Context) {
	var newUser model.UserLogin
	var err error

	if err = context.BindJSON(&newUser); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "неверный формат данных о новом пользователе"})
		return
	}

	if newUser.Password, err = h.hasher.Hash(newUser.Password); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "неверный формат пароля"})
		return
	}

	if _, err = h.repo.CreateUser(context.Request.Context(), newUser); err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "нельзя создать пользователя с таким логином"})

		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "пользователь успешно создан"})
}

// Logout godoc
// @Summary      Logout
// @Description  Logs out the user by blacklisting the access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      400
// @Router       /api/logout [post]
func (h *Handler) Logout(context *gin.Context) {
	jwtStr, err := context.Cookie("AccessToken")
	if !strings.HasPrefix(jwtStr, model.JwtPrefix) || err != nil { // если нет префикса то нас дурят!
		context.AbortWithStatus(http.StatusBadRequest) // отдаем что нет доступа
		return
	}

	// отрезаем префикс
	jwtStr = jwtStr[len(model.JwtPrefix):]

	_, err = h.tokenManager.Parse(jwtStr)
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)
		return
	}

	// сохраняем в блеклист редиса
	err = h.redisClient.WriteJWTToBlacklist(context.Request.Context(), jwtStr, time.Hour)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	context.Status(http.StatusOK)
}
