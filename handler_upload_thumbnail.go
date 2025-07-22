package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	const maxMemory = 10 << 20 // This is 20 MiB as max data size
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing the multipart form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()
	mediaType := header.Header.Get("Content-Type")
	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unexpected error", err)
		return
	}
	videoInfo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Video not found", err)
		return
	}
	if videoInfo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", errors.New(""))
		return
	}

	videoThumbnails[videoID] = thumbnail{
		data:      data,
		mediaType: mediaType,
	}

	thumbnailUrl := fmt.Sprintf("http://localhost:8091/api/thumbnails/%s", videoID)

	updatedVideoInfo := database.Video{
		ID:                videoInfo.ID,
		CreatedAt:         videoInfo.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      &thumbnailUrl,
		CreateVideoParams: videoInfo.CreateVideoParams,
	}

	err = cfg.db.UpdateVideo(updatedVideoInfo)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Video not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedVideoInfo)
}
