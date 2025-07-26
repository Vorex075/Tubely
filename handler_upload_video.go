package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const maxVideoSize = (1 << 30)
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

	fmt.Println("uploading video data", videoID, "by user", userID)

	err = r.ParseMultipartForm(maxVideoSize)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing the multipart form", err)
		return
	}

	videoInfo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Video ID does not exist", err)
		return
	}

	if videoInfo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	file, fileHeader, err := r.FormFile("video")
	defer file.Close()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing the multipart form", err)
		return
	}

	mediaType := fileHeader.Header.Get("Content-Type")
	mediaType, _, err = mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unrecognized media type", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Invalid media type", fmt.Errorf("using '%s' as media type, when 'video/mp4' is required", mediaType))
		return
	}

	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating temp file", err)
		return
	}

	_, err = io.Copy(tempFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server file error", err)
		return
	}

	if _, err = tempFile.Seek(0, 0); err != nil {
		respondWithError(w, http.StatusInternalServerError, "seek error", err)
		return
	}

	randomData := make([]byte, 32)
	_, err = rand.Read(randomData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server file error", err)
		return
	}
	objectKey := hex.EncodeToString(randomData) + ".mp4"

	objectParam := s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &objectKey,
		Body:        tempFile,
		ContentType: &mediaType,
	}
	_, err = cfg.s3Client.PutObject(r.Context(), &objectParam)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error from s3", err)
		return
	}

	videoUrl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		cfg.s3Bucket, cfg.s3Region, objectKey)

	newVideoInfo := videoInfo

	newVideoInfo.UpdatedAt = time.Now()
	newVideoInfo.VideoURL = &videoUrl
	err = cfg.db.UpdateVideo(newVideoInfo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unexpected db error", err)
		return
	}
	respondWithJSON(w, http.StatusOK, struct{}{})

}
