package internal

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"tiflo/pkg/redis"
	"time"

	"tiflo/model"
	"tiflo/pkg/auth"
	"tiflo/pkg/grpc/client"
	pythonClient "tiflo/pkg/grpc/client"
	pb "tiflo/pkg/grpc/generated"
	"tiflo/pkg/hash"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	PathForMedia = "/media/"
)

type Handler struct {
	logger *logrus.Entry

	redisClient redis.Client
	repo        Repository

	hasher       hash.PasswordHasher
	tokenManager auth.TokenManager
	pythonClient client.AI
}

func initConfig(vp *viper.Viper, configPath string) error {
	path := filepath.Dir(configPath)
	vp.AddConfigPath(path)
	vp.SetConfigName(strings.Split(filepath.Base(configPath), ".")[0])
	vp.SetConfigType(filepath.Ext(configPath)[1:])
	return vp.ReadInConfig()
}

func parseFlags() (*bool, *bool) {
	python := flag.Bool("python", true, "use python for ai, otherwise use mock for models")
	db := flag.Bool("db", true, "use postgres as db")

	flag.Parse()

	return python, db
}

func NewHandler(logger *logrus.Logger) *Handler {
	vp := viper.New()
	if err := initConfig(vp, "/configs/config.yml"); err != nil {
		logger.Printf("error initializing configs: %s\n", err.Error())
	}

	pythonNeeded, dbNeeded := parseFlags()

	var repos Repository

	if *dbNeeded {
		db, err := NewPostgresDB(vp.GetString("db.connection_string"))
		if err != nil {
			log.Fatal("error during connecting to postgres ", err)
		}
		logger.Info("connected to postgres")

		repos = NewRepository(logger, db)
	}

	var pythonCl *pythonClient.PythonClient
	if *pythonNeeded {
		address := vp.GetString("python.address")
		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		logger.Info("connected to python")

		client := pb.NewAIServiceClient(conn)
		pythonCl = pythonClient.NewPythonClient(logger, client)
	}

	tokenManager, err := auth.NewManager(vp.GetString("auth.secret"))
	if err != nil {
		logger.Fatalln(err)
	}

	redisConfig := redis.InitRedisConfig(vp)

	redisClient, err := redis.NewRedisClient(context.Background(), redisConfig, logger)
	if err != nil {
		logger.Fatalln(err)
	}

	return &Handler{
		logger:       logger.WithField("component", "handler"),
		pythonClient: pythonCl,
		repo:         repos,
		hasher:       hash.NewSHA256Hasher(vp.GetString("auth.salt")),
		tokenManager: tokenManager,
		redisClient:  redisClient,
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (h *Handler) InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(h.AuthCheck())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "pong")
	})

	apiGroup := r.Group("/api")
	{
		authRouter := apiGroup.Group("/auth")
		{
			authRouter.POST("/signIn", h.SignIn)
			authRouter.POST("/signUp", h.SignUp)
			authRouter.POST("/logout", h.Logout)
		}

	}

	//apiGroup.POST("/save-image", h.SaveImage)
	//apiGroup.GET("/get-image-project", h.GetImageProject)

	return r
}

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

//
//func (h *Handler) SaveImage(context *gin.Context) {
//	form, _ := context.MultipartForm()
//	files := form.File["file"]
//
//	for _, file := range files {
//		if err := context.SaveUploadedFile(file, PathForMedia+file.Filename); err != nil {
//			h.logger.Printf("Failed to save file: %s", err)
//			context.String(500, "Failed to save file")
//			return
//		}
//
//		if err := h.repo.SaveImageProject(context.Request.Context(), model.ImageProject{Image: file.Filename}); err != nil {
//			context.String(500, "Failed to save file")
//			return
//		}
//	}
//
//	context.String(200, "File uploaded successfully")
//}
//
//func (h *Handler) GetImageProject(context *gin.Context) {
//	project, err := h.repo.GetImageProject(context.Request.Context(), model.ImageProject{})
//	if err != nil {
//		context.String(500, err.Error())
//		return
//	}
//
//	context.JSON(http.StatusOK, project)
//}
//
//func (h *Handler) GetVoicedText(context *gin.Context) {
//	var textComment model.TextToVoice
//	h.logger.Info(textComment)
//	if err := context.BindJSON(&textComment); err != nil {
//		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
//		return
//	}
//
//	audioBytes, err := h.pythonClient.VoiceTheText(context.Request.Context(), textComment)
//	if err != nil {
//		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
//		return
//	}
//
//	err = ioutil.WriteFile("test.wav", audioBytes, 0644)
//	if err != nil {
//		log.Fatalf("Failed to write file: %s", err)
//	}
//
//	context.Writer.Header().Set("Content-Type", "audio/wav")
//	_, err = context.Writer.Write(audioBytes)
//	if err != nil {
//		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
//		return
//	}
//}
