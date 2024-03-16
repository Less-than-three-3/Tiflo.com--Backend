package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"strings"
	"tiflo/internal"
	"time"
)

func main() {
	logger := logrus.New()
	formatter := &logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	}
	logger.SetFormatter(formatter)

	//addr := "172.21.133.141:8080"
	//conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer conn.Close()
	//logger.Info("connected to python")

	vp := viper.New()
	if err := initConfig(vp, "/configs/config.yml"); err != nil {
		log.Printf("error initializing configs: %s\n", err.Error())
	}

	db, err := internal.NewPostgresDB(vp.GetString("db.connection_string"))
	if err != nil {
		log.Fatal("error during connecting to postgres ", err)
	}
	logger.Info("connected to postgres")

	repos := internal.NewRepository(logger, db)
	// client := pb.NewAIServiceClient(conn)
	//handler := internal.NewHandler(logger, pythonClient.NewPythonClient(logger, client), repos)
	handler := internal.NewHandler(logger, repos)

	r := handler.InitRouter()
	r.Run("0.0.0.0:8080")
}

func initConfig(vp *viper.Viper, configPath string) error {
	path := filepath.Dir(configPath)
	vp.AddConfigPath(path)
	vp.SetConfigName(strings.Split(filepath.Base(configPath), ".")[0])
	vp.SetConfigType(filepath.Ext(configPath)[1:])
	return vp.ReadInConfig()
}
