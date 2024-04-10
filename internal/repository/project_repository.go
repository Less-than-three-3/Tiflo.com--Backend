package repository

import (
	"context"
	"database/sql"
	"errors"

	"tiflo/model"

	"github.com/google/uuid"
)

func (r *RepositoryPostgres) CreateProject(context context.Context, userId uuid.UUID) (model.Project, error) {
	query := `INSERT INTO "project"(user_id) VALUES ($1) RETURNING project_id, name, user_id;`
	var newProject model.Project

	row := r.db.QueryRow(context, query, userId)
	if err := row.Scan(&newProject.ProjectId, &newProject.Name, &newProject.UserId); err != nil {
		r.logger.Error(err)
		return model.Project{}, err
	}

	return newProject, nil
}

func (r *RepositoryPostgres) RenameProject(context context.Context, project model.Project) error {
	query := `UPDATE "project" SET name=$1 WHERE project_id=$2 AND user_id=$3;`
	var newProject model.Project

	row := r.db.QueryRow(context, query, project.Name, project.ProjectId, project.UserId)
	if err := row.Scan(&newProject.ProjectId, &newProject.Name, &newProject.UserId); err != nil {
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *RepositoryPostgres) UploadMedia(context context.Context, project model.Project) error {
	query := `UPDATE "project" SET path=$1 WHERE user_id=$2 AND project_id=$3;`

	row := r.db.QueryRow(context, query, project.Path, project.UserId, project.ProjectId)
	if err := row.Scan(); err != nil && !errors.Is(sql.ErrNoRows, err) {
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *RepositoryPostgres) DeleteProject(context context.Context, project model.Project) error {
	query := `DELETE FROM project WHERE project_id=$1 AND user_id=$2;`

	row := r.db.QueryRow(context, query, project.ProjectId, project.UserId)
	if err := row.Scan(); err != nil && !errors.Is(sql.ErrNoRows, err) {
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *RepositoryPostgres) GetProject(context context.Context, project model.Project) (model.Project, error) {
	query := `
	SELECT 
		p.name,
		p.path,
		ap.part_id,
		ap.start,
		ap.duration,
		ap.text,
		ap.path
	FROM 
		project p
	LEFT JOIN 
		audio_part ap ON p.project_id = ap.project_id
	WHERE 
		p.project_id = $1 AND user_id=$2
	`

	rows, err := r.db.Query(context, query, project.ProjectId, project.UserId)
	if err != nil {
		return model.Project{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var ap model.AudioPart
		err = rows.Scan(&project.Name, &project.Path, &ap.PartId, &ap.Start, &ap.Duration, &ap.Text, &ap.Path)
		if err != nil {
			return model.Project{}, err
		}
		project.AudioParts = append(project.AudioParts, ap)
	}

	if err = rows.Err(); err != nil {
		return model.Project{}, err
	}

	return project, nil
}

func (r *RepositoryPostgres) GetProjectsList(context context.Context, userId uuid.UUID) ([]model.Project, error) {
	query := `
	SELECT 
		p.project_id,
		p.name,
		p.path,
		p.user_id,
		ap.part_id,
		ap.start,
		ap.duration,
		ap.text,
		ap.path
	FROM 
		project p
	LEFT JOIN 
		audio_part ap ON p.project_id = ap.project_id
	WHERE 
		p.user_id = $1
	`
	// TODO check if pgxpool support array_ag for group by constructions
	rows, err := r.db.Query(context, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := map[uuid.UUID]model.Project{}
	for rows.Next() {
		var projectId uuid.UUID
		var name, path string
		var userId uuid.UUID
		var partId uuid.UUID
		var start, duration int64
		var text, partPath string

		err = rows.Scan(&projectId, &name, &path, &userId, &partId, &start, &duration, &text, &partPath)
		if err != nil {
			return nil, err
		}

		project, exists := projects[projectId]
		if !exists {
			project = model.Project{
				ProjectId:  projectId,
				Name:       name,
				Path:       path,
				UserId:     userId,
				AudioParts: []model.AudioPart{},
			}
			projects[projectId] = project
		}

		project.AudioParts = append(project.AudioParts, model.AudioPart{
			PartId:   partId,
			Start:    start,
			Duration: duration,
			Text:     text,
			Path:     partPath,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	var result []model.Project
	for _, project := range projects {
		result = append(result, project)
	}

	return result, nil
}
