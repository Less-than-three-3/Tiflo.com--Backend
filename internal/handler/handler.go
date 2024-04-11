package handler

import (
	"context"
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"tiflo/internal/repository"

	_ "tiflo/docs"
	"tiflo/pkg/auth"
	"tiflo/pkg/grpc/client"
	pythonClient "tiflo/pkg/grpc/client"
	pb "tiflo/pkg/grpc/generated"
	"tiflo/pkg/hash"
	"tiflo/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

type Handler struct {
	logger *logrus.Entry

	redisClient redis.Client
	repo        repository.Repository

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

	var repos repository.Repository

	if *dbNeeded {
		db, err := repository.NewPostgresDB(vp.GetString("db.connection_string"))
		if err != nil {
			log.Fatal("error during connecting to postgres ", err)
		}
		logger.Info("connected to postgres")

		repos = repository.NewRepository(logger, db)
	}

	var pythonCl *pythonClient.PythonClient
	if *pythonNeeded {
		voice2textAddress := vp.GetString("python.voice2text.address")
		conn, err := grpc.Dial(voice2textAddress, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatal(err)
		}

		logger.Info("connected to voice2text")

		voice2textClient := pb.NewAIServiceClient(conn)

		image2textAddress := vp.GetString("python.image2text.address")
		conn, err = grpc.Dial(image2textAddress, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatal(err)
		}

		logger.Info("connected to image2text")

		image2textClient := pb.NewImageCaptioningClient(conn)

		pythonCl = pythonClient.NewPythonClient(logger, voice2textClient, image2textClient)
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
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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

		routerWithAuthCheck := apiGroup.Group("/")
		routerWithAuthCheck.Use(h.AuthCheck())

		projectsRouter := routerWithAuthCheck.Group("/projects")
		{
			projectsRouter.POST("/", h.CreateProject)
			projectsRouter.GET("/", h.GetProjects)
			projectsRouter.PATCH("/:projectId/", h.UpdateProjectName)
			projectsRouter.DELETE("/:projectId/", h.DeleteProject)
			projectsRouter.GET("/:projectId/", h.GetProjectInfo)

			projectsRouter.POST("/:projectId/media", h.UploadMedia)
			//projectsRouter.POST("/:projectId/tifloComment/image", h.AddTifloCommentToImage)

			projectsRouter.POST("/:projectId/voice", h.VoiceText)
			projectsRouter.POST("/:projectId/image/text", h.ImageToText)
		}

	}

	return r
}
