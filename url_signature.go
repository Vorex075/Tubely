package main

/*import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)*/

/*func generatePresignedURL(s3Client *s3.Client,
	bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	objectParam := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	presignedHTTP, err := presignClient.PresignGetObject(context.Background(),
		&objectParam,
		s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}
	return presignedHTTP.URL, nil
}*/

/*func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	keys := strings.Split(*video.VideoURL, ",")
	bucket := keys[0]
	videoKey := keys[1]
	signedURL, err := generatePresignedURL(cfg.s3Client, bucket, videoKey, 100*time.Second)
	if err != nil {
		return database.Video{}, err
	}
	video.VideoURL = &signedURL
	return video, nil
}*/
