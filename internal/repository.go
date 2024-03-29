package internal

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"log"
	"tiflo/model"
)

type Repository interface {
	CreateUser(context context.Context, newUser model.UserLogin) (model.User, error)
	GetUser(context context.Context, user model.UserLogin) (model.User, error)
}

type RepositoryPostgres struct {
	db     *pgxpool.Pool
	logger *logrus.Entry
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

func (r *RepositoryPostgres) CreateUser(context context.Context, newUser model.UserLogin) (model.User, error) {
	query := `INSERT INTO "users"(login, password) VALUES ($1, $2) RETURNING user_id;`
	var newUserInfo = model.User{Login: newUser.Login}

	row := r.db.QueryRow(context, query, newUser.Login, newUser.Password)
	if err := row.Scan(&newUserInfo.UserId); err != nil {
		r.logger.Error(err)
		if pqError, ok := err.(*pgconn.PgError); ok {
			if pqError.Code == "23505" {
				return model.User{}, model.Conflict
			}
		}

		return model.User{}, err
	}

	return newUserInfo, nil
}

func (r *RepositoryPostgres) GetUser(context context.Context, user model.UserLogin) (model.User, error) {
	var userInfo model.User
	query := `SELECT user_id, login FROM "users" WHERE login=$1 AND password=$2;`

	row := r.db.QueryRow(context, query, user.Login, user.Password)
	if err := row.Scan(&userInfo.UserId, userInfo.Login); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return model.User{}, model.NotFound
		}
		r.logger.Error(err)
		return model.User{}, err
	}

	return userInfo, nil
}

//
//func (r *RepositoryPostgres) SaveImageProject(context context.Context, project model.ImageProject) error {
//	query := `INSERT INTO project(image_name, project_name, user_id, project_id) VALUES ($1, $2, $3, $4) RETURNING project_id;`
//	var projectId uuid.UUID
//
//	project.UserId = GetUserId()
//	project.Name = "awesomeProject"
//	project.ProjectId = GetProjectId()
//	row := r.db.QueryRow(context, query, project.Image, project.Name, project.UserId, project.ProjectId)
//	if err := row.Scan(&projectId); err != nil {
//		r.logger.Error(err)
//		return err
//	}
//
//	return nil
//}

//func (r *RepositoryPostgres) GetImageProject(context context.Context, project model.ImageProject) (model.ImageProject, error) {
//	var projectInfo model.ImageProject
//	query := `SELECT user_id, project_name, image_name, project_id FROM project WHERE project_id=$1;`
//
//	row := r.db.QueryRow(context, query, GetProjectId())
//	if err := row.Scan(&projectInfo.UserId, &projectInfo.Name, &projectInfo.Image); err != nil {
//		r.logger.Error(err)
//		return model.ImageProject{}, err
//	}
//
//	return projectInfo, nil
//}
