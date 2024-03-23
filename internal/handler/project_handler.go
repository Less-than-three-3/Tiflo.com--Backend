package handler

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
