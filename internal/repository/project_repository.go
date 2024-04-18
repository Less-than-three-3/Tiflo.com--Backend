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
	query := `UPDATE "project" SET video_path=$1, audio_path=$2 WHERE user_id=$3 AND project_id=$4 RETURNING video_path;`

	var path string
	row := r.db.QueryRow(context, query, project.VideoPath, project.AudioPath, project.UserId, project.ProjectId)
	if err := row.Scan(&path); err != nil {
		r.logger.Error(err)
		return err
	}

	if len(project.AudioParts) > 0 {
		var projectId uuid.UUID
		query2 := `INSERT INTO "audio_part"(part_id, project_id, path) VALUES ($1, $2, $3) RETURNING project_id;`
		row = r.db.QueryRow(context, query2, project.AudioParts[0].PartId, project.AudioParts[0].ProjectId, project.AudioParts[0].Path)
		if err := row.Scan(&projectId); err != nil {
			r.logger.Error(err)
			return err
		}
	}

	return nil
}

func (r *RepositoryPostgres) SaveProjectAudio(context context.Context, project model.Project) error {
	var projectId uuid.UUID
	query := `DELETE FROM "audio_part" WHERE project_id=$1 RETURNING project_id;`
	row := r.db.QueryRow(context, query, project.ProjectId)
	if err := row.Scan(&projectId); err != nil {
		r.logger.Error(err)
		return err
	}

	for _, v := range project.AudioParts {
		query = `INSERT INTO "audio_part"(part_id, project_id, start, text, path, duration) VALUES ($1, $2, $3, $4, $5, $6) RETURNING project_id;`
		row = r.db.QueryRow(context, query, v.PartId, v.ProjectId, v.Start, v.Text, v.Path, v.Duration)
		if err := row.Scan(&projectId); err != nil {
			r.logger.Error(err)
			return err
		}
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
		p.video_path,
		p.audio_path,
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

	var projectVideoPath, projectAudioPath sql.NullString
	for rows.Next() {
		var ap model.AudioPart
		var audioPath, audioText sql.NullString
		var duration, start sql.NullInt64

		err = rows.Scan(&project.Name, &projectVideoPath, &projectAudioPath, &ap.PartId, &start, &duration, &audioText, &audioPath)
		if err != nil {
			return model.Project{}, err
		}
		project.VideoPath = projectVideoPath.String
		project.AudioPath = projectAudioPath.String
		ap.Path = audioPath.String
		ap.Start = start.Int64
		ap.Duration = duration.Int64

		ap.Text = audioText.String

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
		p.video_path,
		p.audio_path,
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
		var name string
		var userId uuid.UUID
		var partId uuid.UUID
		var projectPath sql.NullString
		var projectAudioPath sql.NullString

		var audioPath, audioText sql.NullString
		var duration, start sql.NullInt64

		err = rows.Scan(&projectId, &name, &projectPath, &projectAudioPath, &userId, &partId, &start, &duration, &audioText, &audioPath)
		if err != nil {
			return nil, err
		}

		project, exists := projects[projectId]
		if !exists {
			project = model.Project{
				ProjectId:  projectId,
				Name:       name,
				VideoPath:  projectPath.String,
				AudioPath:  projectAudioPath.String,
				UserId:     userId,
				AudioParts: []model.AudioPart{},
			}
			projects[projectId] = project
		}

		project.AudioParts = append(project.AudioParts, model.AudioPart{
			PartId:   partId,
			Start:    start.Int64,
			Duration: duration.Int64,
			Text:     audioText.String,
			Path:     audioPath.String,
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
