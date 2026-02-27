package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	SupabaseURL    string
	SupabaseKey    string
	SupabaseBucket string
)

/*
	InitSupabase inicializa variáveis do Supabase Storage (REST API)
	⚠️ Use SEMPRE a SERVICE_ROLE_KEY no backend
*/
func InitSupabase() {
	SupabaseURL = strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	SupabaseKey = strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
	SupabaseBucket = strings.TrimSpace(os.Getenv("SUPABASE_BUCKET"))

	if SupabaseURL == "" || SupabaseKey == "" || SupabaseBucket == "" {
		panic("❌ SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY e SUPABASE_BUCKET são obrigatórios")
	}

	fmt.Println("✅ Supabase Storage (REST) inicializado com sucesso")
}

// IsValidImage valida tipos de imagem permitidos
func IsValidImage(file *multipart.FileHeader) bool {
	switch file.Header.Get("Content-Type") {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

/*
	UploadFile envia arquivo para o Supabase Storage
	Path final: <petID>/<slug-pet>-<posição>.<extensão>
*/
func UploadFile(
	file *multipart.FileHeader,
	petID string,
	petName string,
	position int,
) (string, error) {

	if !IsValidImage(file) {
		return "", fmt.Errorf("tipo de imagem não suportado: %s", file.Header.Get("Content-Type"))
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	objectPath := fmt.Sprintf(
		"%s/%s-%d%s",
		petID,
		slugify(petName),
		position,
		ext,
	)

	// 🔹 Monta multipart body manualmente (com MIME correto)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	header := make(textproto.MIMEHeader)
	header.Set(
		"Content-Disposition",
		fmt.Sprintf(`form-data; name="file"; filename="%s"`, file.Filename),
	)
	header.Set("Content-Type", file.Header.Get("Content-Type"))

	part, err := writer.CreatePart(header)
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(part, src); err != nil {
		return "", err
	}

	if err = writer.Close(); err != nil {
		return "", err
	}

	uploadURL := fmt.Sprintf(
		"%s/storage/v1/object/%s/%s",
		SupabaseURL,
		SupabaseBucket,
		objectPath,
	)

	req, err := http.NewRequest(http.MethodPost, uploadURL, body)
	if err != nil {
		return "", err
	}

	// 🔥 HEADERS OBRIGATÓRIOS DO SUPABASE
	req.Header.Set("Authorization", "Bearer "+SupabaseKey)
	req.Header.Set("apikey", SupabaseKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-upsert", "false")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"❌ upload falhou (%d): %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	publicURL := fmt.Sprintf(
		"%s/storage/v1/object/public/%s/%s",
		SupabaseURL,
		SupabaseBucket,
		objectPath,
	)

	return publicURL, nil
}

// DeleteFile remove arquivo do bucket
func DeleteFile(objectPath string) error {
	deleteURL := fmt.Sprintf(
		"%s/storage/v1/object/%s/%s",
		SupabaseURL,
		SupabaseBucket,
		objectPath,
	)

	req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+SupabaseKey)
	req.Header.Set("apikey", SupabaseKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"❌ delete falhou (%d): %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	return nil
}

// slugify transforma texto em slug seguro
func slugify(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = strings.ReplaceAll(s, " ", "-")

	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	return reg.ReplaceAllString(s, "")
}