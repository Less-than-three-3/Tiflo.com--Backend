package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
	"tiflo/model"

	"github.com/google/uuid"
)

func (r *RepositoryPostgres) CreateProject(context context.Context, userId uuid.UUID) (model.Project, error) {
	query := `INSERT INTO "project"(user_id) VALUES ($1) RETURNING project_id, name, user_id, created;`
	var newProject model.Project

	row := r.db.QueryRow(context, query, userId)
	if err := row.Scan(&newProject.ProjectId, &newProject.Name, &newProject.UserId, &newProject.Created); err != nil {
		r.logger.Error(err)
		return model.Project{}, err
	}

	return newProject, nil
}

func (r *RepositoryPostgres) UpdateAudioPart(context context.Context, audioPart model.AudioPart) error {
	var partId uuid.UUID
	query := `INSERT INTO "audio_part" (part_id, project_id, start, duration, text, path)
			VALUES
    		($1, $2, $3, $4, $5, $6)
			ON CONFLICT (part_id) DO UPDATE
			SET start = EXCLUDED.start, 
			    duration = EXCLUDED.duration, 
			    text = EXCLUDED.text,
			    path = EXCLUDED.path
			    RETURNING part_id;
	`

	row := r.db.QueryRow(context, query, audioPart.PartId, audioPart.ProjectId, audioPart.Start,
		audioPart.Duration, audioPart.Text, audioPart.Path)
	if err := row.Scan(&partId); err != nil {
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *RepositoryPostgres) RenameProject(context context.Context, project model.Project) error {
	query := `UPDATE "project" SET name=$1 WHERE project_id=$2 AND user_id=$3 RETURNING project_id, name, user_id;`
	var newProject model.Project

	row := r.db.QueryRow(context, query, project.Name, project.ProjectId, project.UserId)
	if err := row.Scan(&newProject.ProjectId, &newProject.Name, &newProject.UserId); err != nil {
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *RepositoryPostgres) ChangeCommentText(context context.Context, project model.Project) error {
	query := `UPDATE audio_part SET text=$1 WHERE project_id=$2 AND part_id=$3 RETURNING part_id;`
	var newProject model.Project

	row := r.db.QueryRow(context, query, project.Name, project.ProjectId, project.UserId)
	if err := row.Scan(&newProject.ProjectId, &newProject.Name, &newProject.UserId); err != nil {
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *RepositoryPostgres) UploadMedia(context context.Context, project model.Project) error {
	query := `UPDATE "project" SET video_path=$1, audio_path=$2, image_path=$3 WHERE user_id=$4 AND project_id=$5 RETURNING video_path;`

	var path string
	row := r.db.QueryRow(context, query, project.VideoPath, project.AudioPath, project.ImagePath, project.UserId, project.ProjectId)
	if err := row.Scan(&path); err != nil {
		r.logger.Error(err)
		return err
	}

	if len(project.AudioParts) > 0 {
		var projectId uuid.UUID
		query2 := `INSERT INTO "audio_part"(part_id, project_id, path, duration, start) VALUES ($1, $2, $3, $4, 0) RETURNING project_id;`
		row = r.db.QueryRow(context, query2, project.AudioParts[0].PartId, project.AudioParts[0].ProjectId,
			project.AudioParts[0].Path, project.AudioParts[0].Duration)
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
	_ = row.Scan(&projectId)

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
	if err := row.Scan(); err != nil && !errors.Is(pgx.ErrNoRows, err) {
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
		p.image_path,
		p.created,
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
		var created sql.NullTime
		var duration, start sql.NullInt64

		err = rows.Scan(&project.Name, &projectVideoPath, &projectAudioPath, &project.AudioPath, &created, &ap.PartId, &start, &duration, &audioText, &audioPath)
		if err != nil {
			return model.Project{}, err
		}
		project.VideoPath = projectVideoPath.String
		project.AudioPath = projectAudioPath.String
		project.Created = created.Time
		ap.Path = audioPath.String
		ap.Start = start.Int64
		ap.Duration = duration.Int64
		ap.ProjectId = project.ProjectId
		ap.Text = audioText.String

		project.AudioParts = append(project.AudioParts, ap)
	}

	if err = rows.Err(); err != nil {
		return model.Project{}, err
	}

	return project, nil
}

func (r *RepositoryPostgres) GetAudioPart(context context.Context, part model.AudioPart) (model.AudioPart, error) {
	query := `
	SELECT 
		part_id,
		start,
		project_id,
		duration,
		text,
		path
	FROM 
		audio_part WHERE part_id=$1`

	rows, err := r.db.Query(context, query, part.PartId)
	if err != nil {
		return model.AudioPart{}, err
	}
	defer rows.Close()

	row := r.db.QueryRow(context, query, part.PartId)
	if err = row.Scan(&part.PartId, &part.ProjectId, &part.Start, &part.Duration, &part.Path); err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			r.logger.Error(err)
			return model.AudioPart{}, model.NotFound
		}
		return model.AudioPart{}, err
	}

	return part, nil
}

func (r *RepositoryPostgres) GetAudioPartBySplitPoint(context context.Context, splitPoint int64,
	projectId uuid.UUID) (model.AudioPart, error) {
	query := `
	SELECT part_id, project_id, start, duration, path
	FROM audio_part
	WHERE 
		 project_id=$1 AND start < $2 AND (start + duration) > $2;
	`

	var audioPart model.AudioPart
	row := r.db.QueryRow(context, query, projectId, splitPoint)
	if err := row.Scan(&audioPart.PartId, &audioPart.ProjectId, &audioPart.Start, &audioPart.Duration,
		&audioPart.Path); err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			r.logger.Error(err)
			return model.AudioPart{}, model.NotFound
		}
		return model.AudioPart{}, err
	}

	return audioPart, nil
}

func (r *RepositoryPostgres) DeleteAudioPart(context context.Context, audioPart model.AudioPart) (model.AudioPart, error) {
	query := `
	DELETE FROM audio_part WHERE part_id = $1 AND project_id=$2 RETURNING part_id, duration, start, text, path;
	`
	var path, text sql.NullString

	row := r.db.QueryRow(context, query, audioPart.PartId, audioPart.ProjectId)
	if err := row.Scan(&audioPart.PartId, &audioPart.Duration, &audioPart.Start, &text, &path); err != nil {
		r.logger.Error(err)
		return audioPart, err
	}
	audioPart.Text = text.String
	audioPart.Path = path.String

	return audioPart, nil
}

func (r *RepositoryPostgres) GetAudioPartsAfterSplitPoint(context context.Context, splitPoint int64,
	projectId uuid.UUID) ([]model.AudioPart, error) {
	query := `
	SELECT part_id, project_id, start, duration, text, path
	FROM audio_part
	WHERE 
		 project_id=$1 AND start > $2;
	`

	var audioParts []model.AudioPart
	rows, err := r.db.Query(context, query, projectId, splitPoint)
	if err != nil && !errors.Is(pgx.ErrNoRows, err) {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var audioPart model.AudioPart

		if err = rows.Scan(&audioPart.PartId, &audioPart.ProjectId, &audioPart.Start, &audioPart.Duration,
			&audioPart.Text, &audioPart.Path); err != nil {
			r.logger.Error(err)
			return nil, err
		}

		audioParts = append(audioParts, audioPart)
	}

	return audioParts, nil
}

func (r *RepositoryPostgres) GetProjectsList(context context.Context, userId uuid.UUID) ([]model.Project, error) {
	query := `
	SELECT 
		p.project_id,
		p.created,
		p.name,
		p.video_path,
		p.audio_path,
		p.image_path,
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
		var created sql.NullTime
		var partId uuid.UUID
		var projectPath sql.NullString
		var projectAudioPath sql.NullString
		var projectImagePath sql.NullString

		var audioPath, audioText sql.NullString
		var duration, start sql.NullInt64

		err = rows.Scan(&projectId, &created, &name, &projectPath, &projectAudioPath, &projectImagePath, &userId, &partId, &start, &duration, &audioText, &audioPath)
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
				ImagePath:  projectImagePath.String,
				UserId:     userId,
				Created:    created.Time,
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
