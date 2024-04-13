package handler

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os/exec"
	"path/filepath"
	"tiflo/model"

	"github.com/gin-gonic/gin"
)

// CreateProject godoc
// @Summary      Create new user project
// @Description  Create a  new project with default name
// @Tags         Project
// @Produce      json
// @Success      200  {object}  model.Project
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/ [post]
func (h *Handler) CreateProject(context *gin.Context) {
	userId, err := model.GetUserId(context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	newProject, err := h.repo.CreateProject(context.Request.Context(), userId)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, newProject)
}

type ProjectUpdate struct {
	Name string `json:"name" binding:"required"`
}

// UpdateProjectName godoc
// @Summary      Update project name
// @Description  rename project
// @Tags         Project
// @Accept       json
// @Produce      json
// @Param        user  body  ProjectUpdate  true  "New project name"
// @Param        projectId  path  string  true  "Project Id"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/{projectId} [patch]
func (h *Handler) UpdateProjectName(context *gin.Context) {
	projectIdStr := context.Param("projectId")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	userId, err := model.GetUserId(context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	if err = h.repo.RenameProject(context.Request.Context(), model.Project{
		ProjectId: projectId,
		UserId:    userId,
	}); err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "проект успешно переименован"})

}

const (
	PathForMedia = "/media/"
)

// UploadMedia godoc
// @Summary      Upload media file for project
// @Description  Uploads a media file to the server
// @Tags         Project
// @Accept       mpfd
// @Produce      json
// @Param        file formData file true "Media file to upload"
// @Param        projectId  path  string  true  "Project Id"
// @Success      200 {object} map[string]any "Successfully uploaded"
// @Failure      500 {object} map[string]any "Failed to save file"
// @Router       /api/projects/{projectId}/media [post]
func (h *Handler) UploadMedia(context *gin.Context) {
	projectIdStr := context.Param("projectId")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	userId, err := model.GetUserId(context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	form, _ := context.MultipartForm()
	files := form.File["file"]
	filename := uuid.New()

	for _, file := range files {
		extension := filepath.Ext(file.Filename)
		if err = context.SaveUploadedFile(file, PathForMedia+filename.String()+extension); err != nil {
			h.logger.Printf("Failed to save file: %s", err)
			context.String(http.StatusInternalServerError, "Failed to save file")
			return
		}

		var audio []model.AudioPart

		if extension == ".mp4" {
			_, err = exec.Command("ffmpeg", "-i", PathForMedia+filename.String()+extension, "-vn", PathForMedia+filename.String()+".mp3").Output()
			if err != nil {
				fmt.Printf("error %s", err)
			}

			audio = append(audio, model.AudioPart{
				PartId:    uuid.New(),
				ProjectId: projectId,
				Start:     0,
				Duration:  0,
				Text:      "",
				Path:      PathForMedia + filename.String() + ".mp3",
			})
		}

		if err = h.repo.UploadMedia(context.Request.Context(), model.Project{
			ProjectId:  projectId,
			VideoPath:  filename.String() + extension,
			AudioPath:  PathForMedia + filename.String() + ".mp3",
			UserId:     userId,
			AudioParts: audio,
		}); err != nil {
			context.String(http.StatusInternalServerError, "Failed to save file")
			return
		}
	}

	context.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

// DeleteProject godoc
// @Summary      Delete project
// @Description  Delete project
// @Tags         Project
// @Produce      json
// @Param        projectId  path  string  true  "Project Id"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/{projectId} [delete]
func (h *Handler) DeleteProject(context *gin.Context) {
	projectIdStr := context.Param("projectId")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	userId, err := model.GetUserId(context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	err = h.repo.DeleteProject(context.Request.Context(), model.Project{ProjectId: projectId, UserId: userId})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Project removed successfully"})
}

// GetProjectInfo godoc
// @Summary      Get project info
// @Description  Get project name, path to media and audio parts
// @Tags         Project
// @Produce      json
// @Param        projectId  path  string  true  "Project Id"
// @Success      200  {object}  model.Project
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/{projectId} [get]
func (h *Handler) GetProjectInfo(context *gin.Context) {
	projectIdStr := context.Param("projectId")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	userId, err := model.GetUserId(context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	project, err := h.repo.GetProject(context.Request.Context(), model.Project{ProjectId: projectId, UserId: userId})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, project)
}

// GetProjects godoc
// @Summary      Get all user' projects info
// @Description  Get all user' projects as an array
// @Tags         Project
// @Produce      json
// @Success      200  {object}  []model.Project
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/ [get]
func (h *Handler) GetProjects(context *gin.Context) {
	userId, err := model.GetUserId(context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	projects, err := h.repo.GetProjectsList(context.Request.Context(), userId)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, projects)
}

//func (h *Handler) AddTifloCommentToImage(context *gin.Context) {
//	projectIdStr := context.Param("projectId")
//	projectId, err := uuid.Parse(projectIdStr)
//	if err != nil {
//		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
//		return
//	}
//
//	userId, err := model.GetUserId(context)
//	if err != nil {
//		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
//		return
//	}
//
//	// h.pythonClient.GetComment()
//
//	context.JSON(http.StatusOK, "success")
//}

func (h *Handler) f(context *gin.Context) {

}
