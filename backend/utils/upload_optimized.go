package utils

import (
	"bytes"
	"fmt"
	"net/http"
)

/*
	UploadOptimizedFile
	Faz upload de uma imagem já processada (resize/compress)
*/
func UploadOptimizedFile(
	buffer *bytes.Buffer,
	contentType string,
	extension string,
	bucket string,
	folderID string,
	name string,
	position int,
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

	req.Header.Set("Authorization", "Bearer "+SupabaseKey)
	req.Header.Set("apikey", SupabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "false")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("falha ao fazer upload da imagem (status %d)", resp.StatusCode)
	}

	publicURL := fmt.Sprintf(
		"%s/storage/v1/object/public/%s/%s",
		SupabaseURL,
		bucket,
		objectPath,
	)

	return publicURL, nil
}