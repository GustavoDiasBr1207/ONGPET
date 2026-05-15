package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// UploadOptimizedFile faz upload de uma imagem já processada (resize/compress).
// Usa x-upsert: true para evitar erro 400 quando o arquivo já existe no mesmo path.
func UploadOptimizedFile(
	buffer *bytes.Buffer,
	contentType string,
	extension string,
	bucket string,
	folderID string,
	name string,
	position int,
	authToken string,
) (string, error) {

	objectPath := fmt.Sprintf(
		"%s/%s-%d%s",
		folderID,
		slugify(name),
		position,
		extension,
	)

	uploadURL := fmt.Sprintf(
		"%s/storage/v1/object/%s/%s",
		SupabaseURL,
		bucket,
		objectPath,
	)

	req, err := http.NewRequest(http.MethodPost, uploadURL, buffer)
	if err != nil {
		return "", err
	}

	bearerToken := authToken
	if bearerToken == "" {
		bearerToken = SupabaseKey
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("apikey", SupabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("falha ao fazer upload da imagem (status %d): %s", resp.StatusCode, string(body))
	}

	publicURL := fmt.Sprintf(
		"%s/storage/v1/object/public/%s/%s",
		SupabaseURL,
		bucket,
		objectPath,
	)

	return publicURL, nil
}