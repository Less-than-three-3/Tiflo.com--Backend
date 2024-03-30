package handler

import (
	"github.com/google/uuid"
	"net/http"
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

//func (h *Handler) SaveImage(context *gin.Context) {
//	form, _ := context.MultipartForm()
//	files := form.File["file"]
//
//	for _, file := range files {
//		if err := context.SaveUploadedFile(file, PathForMedia+file.Filename); err != nil {
//			h.logger.Printf("Failed to save file: %s", err)
//			context.String(500, "Failed to save file")
//			return
//		}
//
//		if err := h.repo.SaveImageProject(context.Request.Context(), model.ImageProject{Image: file.Filename}); err != nil {
//			context.String(500, "Failed to save file")
//			return
//		}
//	}
//
//	context.String(200, "File uploaded successfully")
//}
//
//func (h *Handler) GetImageProject(context *gin.Context) {
//	project, err := h.repo.GetImageProject(context.Request.Context(), model.ImageProject{})
//	if err != nil {
//		context.String(500, err.Error())
//		return
//	}
//
//	context.JSON(http.StatusOK, project)
//}
//
//func (h *Handler) GetVoicedText(context *gin.Context) {
//	var textComment model.TextToVoice
//	h.logger.Info(textComment)
//	if err := context.BindJSON(&textComment); err != nil {
//		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
//		return
//	}
//
//	audioBytes, err := h.pythonClient.VoiceTheText(context.Request.Context(), textComment)
//	if err != nil {
//		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
//		return
//	}
//
//	err = ioutil.WriteFile("test.wav", audioBytes, 0644)
//	if err != nil {
//		log.Fatalf("Failed to write file: %s", err)
//	}
//
//	context.Writer.Header().Set("Content-Type", "audio/wav")
//	_, err = context.Writer.Write(audioBytes)
//	if err != nil {
//		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
//		return
//	}
//}
