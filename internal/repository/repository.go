package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"tiflo/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	CreateUser(context context.Context, newUser model.UserLogin) (model.User, error)
	GetUser(context context.Context, user model.UserLogin) (model.User, error)

	CreateProject(context context.Context, userId uuid.UUID) (model.Project, error)
	RenameProject(context context.Context, project model.Project) error
	DeleteProject(context context.Context, project model.Project) error
	GetProject(context context.Context, project model.Project) (model.Project, error)
	GetProjectsList(context context.Context, userId uuid.UUID) ([]model.Project, error)

	UploadMedia(context context.Context, project model.Project) error

	SaveProjectAudio(context context.Context, project model.Project) error

	GetAudioPartBySplitPoint(context context.Context, splitPoint int64, projectId uuid.UUID) (model.AudioPart, error)
	GetAudioPartsAfterSplitPoint(context context.Context, splitPoint int64, projectId uuid.UUID) ([]model.AudioPart, error)
	UpdateAudioParts(context context.Context, audioPart model.AudioPart) error
	DeleteAudioPart(context context.Context, audioPart model.AudioPart) error
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
	query := `INSERT INTO "user"(login, password_hash) VALUES ($1, $2) RETURNING user_id;`
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
	query := `SELECT user_id, login FROM "user" WHERE login=$1 AND password_hash=$2;`

	row := r.db.QueryRow(context, query, user.Login, user.Password)
	if err := row.Scan(&userInfo.UserId, &userInfo.Login); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return model.User{}, model.NotFound
		}
		r.logger.Error(err)
		return model.User{}, err
	}

	return userInfo, nil
}

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
