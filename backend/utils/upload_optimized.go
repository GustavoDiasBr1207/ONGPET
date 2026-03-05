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
	petID string,
	petName string,
	position int,
) (string, error) {

	objectPath := fmt.Sprintf(
		"%s/%s-%d%s",
		petID,
		slugify(petName),
		position,
		extension,
	)

	uploadURL := fmt.Sprintf(
		"%s/storage/v1/object/%s/%s",
		SupabaseURL,
		SupabaseBucket,
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
		SupabaseBucket,
		objectPath,
	)

	return publicURL, nil
}