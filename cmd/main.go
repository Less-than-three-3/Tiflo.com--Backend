package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"
	"tiflo/internal"
	"time"

	pythonClient "tiflo/pkg/grpc/client"
	pb "tiflo/pkg/grpc/generated"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

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

func main() {
	logger := logrus.New()
	formatter := &logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	}
	logger.SetFormatter(formatter)

	vp := viper.New()
	if err := initConfig(vp, "/configs/config.yml"); err != nil {
		logger.Printf("error initializing configs: %s\n", err.Error())
	}

	pythonNeeded, dbNeeded := parseFlags()

	var repos internal.Repository

	if *dbNeeded {
		db, err := internal.NewPostgresDB(vp.GetString("db.connection_string"))
		if err != nil {
			log.Fatal("error during connecting to postgres ", err)
		}
		logger.Info("connected to postgres")

		repos = internal.NewRepository(logger, db)
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

	handler := internal.NewHandler(logger, pythonCl, repos)

	r := handler.InitRouter()
	r.Run("0.0.0.0:8080")
}
