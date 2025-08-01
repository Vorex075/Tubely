package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

var allowedThumbnailMediaType = []string{"image/jpeg", "image/png"}

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
	mediaType, _, err = mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unrecognized mediaType", err)
		return
	}

	if slices.Index(allowedThumbnailMediaType, mediaType) == -1 {
		respondWithError(w, http.StatusBadRequest, "Invalid media type", err)
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

	thumbnailFileExtension := strings.Split(mediaType, "/")[1]
	randomData := make([]byte, 32)
	_, err = rand.Read(randomData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error", err)
		return
	}

	thumbnailFileName := fmt.Sprintf("%s.%s",
		base64.RawURLEncoding.EncodeToString(randomData), thumbnailFileExtension)
	thumbnailNewFile, err := os.Create(path.Join(cfg.assetsRoot, thumbnailFileName))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "File creation error", err)
		return
	}

	_, err = io.Copy(thumbnailNewFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "file creation error", err)
		return
	}

	thumbnailUrl := fmt.Sprintf("http://localhost:8091/assets/%s", thumbnailFileName)

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
