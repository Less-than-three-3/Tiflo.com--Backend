package internal

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"log"
	"tiflo/model"
)

type Repository interface {
	SaveImageProject(context context.Context, project model.ImageProject) error
	GetImageProject(context context.Context, project model.ImageProject) (model.ImageProject, error)
}

type RepositoryPostgres struct {
	db     *pgxpool.Pool
	logger *logrus.Entry
}

func GetProjectId() uuid.UUID {
	uuid, _ := uuid.Parse("4b674f89-f2b5-4935-839f-8977ff76ee38")
	return uuid
}

func GetUserId() uuid.UUID {
	uuid, _ := uuid.Parse("3a964e4e-6a07-4730-b792-06b00c4284c3")
	return uuid
}

func NewPostgresDB(str string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), str)
	if err != nil {
		log.Println("Error while connecting to DB", err)
		return nil, err
	}

	err = db.Ping(context.Background())
	if err != nil {
		log.Println("Error while ping to DB", err)
		return nil, err
	}
	log.Println("connected to postgres")
	return db, err
}

func NewRepository(logger *logrus.Logger, db *pgxpool.Pool) Repository {
	return &RepositoryPostgres{
		db:     db,
		logger: logger.WithField("component", "repo"),
	}
}

func (r *RepositoryPostgres) SaveImageProject(context context.Context, project model.ImageProject) error {
	query := `INSERT INTO project(image_name, project_name, user_id, project_id) VALUES ($1, $2, $3, $4) RETURNING project_id;`
	var projectId uuid.UUID

	project.UserId = GetUserId()
	project.Name = "awesomeProject"
	project.ProjectId = GetProjectId()
	row := r.db.QueryRow(context, query, project.Image, project.Name, project.UserId, project.ProjectId)
	if err := row.Scan(&projectId); err != nil {
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *RepositoryPostgres) GetImageProject(context context.Context, project model.ImageProject) (model.ImageProject, error) {
	var projectInfo model.ImageProject
	query := `SELECT user_id, project_name, image_name, project_id FROM project WHERE project_id=$1;`

	row := r.db.QueryRow(context, query, GetProjectId())
	if err := row.Scan(&projectInfo.UserId, &projectInfo.Name, &projectInfo.Image); err != nil {
		r.logger.Error(err)
		return model.ImageProject{}, err
	}

	return projectInfo, nil
}
