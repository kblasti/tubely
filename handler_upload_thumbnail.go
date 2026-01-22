package main

import (
	"fmt"
	"net/http"
	"io"
	"path/filepath"
	"os"
	"strings"
	"mime"
	"crypto/rand"
	"encoding/base64"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
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

	const maxMemory int64 = 10 << 20

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing memory", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve file", err)
		return
	}

	contentType := header.Header.Get("Content-Type")

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve video", err)
		return
	}
	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "User not video owner", err)
		return
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find media type", err)
		return
	}

	if mediaType != "image/jpeg" {
		if mediaType != "image/png" {
			respondWithError(w, http.StatusBadRequest, "Invalid file type", err)
			return
		}
	}

	ext := ""
	parts := strings.Split(contentType, "/")
	if len(parts) == 2 {
		ext = parts[1]
	}
	
	randKey := make([]byte, 32)
	rand.Read(randKey)
	randURL := base64.RawURLEncoding.EncodeToString(randKey)

	filename := randURL + "." + ext

	path := filepath.Join(cfg.assetsRoot, filename)

	thumbnail, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create filepath", err)
		return
	}
	defer thumbnail.Close()
	defer file.Close()

	_, err = io.Copy(thumbnail, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save data to disk", err)
		return
	}

	thumbnailURL := "http://localhost:" + cfg.port + "/assets/" + filename

	video.ThumbnailURL = &thumbnailURL
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video data", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
