package utils

import (
	"bytes"
	"encoding/json"
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
	SupabaseURL               string
	SupabaseKey               string
	SupabaseBucketPets        string
	SupabaseBucketOngs        string
	SupabaseBucketFormularios string
)

// InitSupabase inicializa variáveis do Supabase Storage (REST API)
func InitSupabase() {
	SupabaseURL = strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	SupabaseKey = strings.TrimSpace(os.Getenv("SUPABASE_KEY")) // ← anon key

	SupabaseBucketPets        = strings.TrimSpace(os.Getenv("SUPABASE_BUCKET_PETS"))
	SupabaseBucketOngs        = strings.TrimSpace(os.Getenv("SUPABASE_BUCKET_ONGS"))
	SupabaseBucketFormularios = strings.TrimSpace(os.Getenv("SUPABASE_BUCKET_FORMULARIOS"))

	if SupabaseURL == "" || SupabaseKey == "" {
		panic("❌ SUPABASE_URL e SUPABASE_KEY são obrigatórios")
	}
	if SupabaseBucketPets == "" || SupabaseBucketOngs == "" || SupabaseBucketFormularios == "" {
		panic("❌ SUPABASE_BUCKET_PETS, SUPABASE_BUCKET_ONGS e SUPABASE_BUCKET_FORMULARIOS são obrigatórios")
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

// UploadFile envia arquivo para o Supabase Storage
func UploadFile(
	file *multipart.FileHeader,
	bucket string,
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
		bucket,
		objectPath,
	)

	req, err := http.NewRequest(http.MethodPost, uploadURL, body)
	if err != nil {
		return "", err
	}

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
		bucket,
		objectPath,
	)

	return publicURL, nil
}

func DeleteFile(bucket string, objectPath string) error {
	deleteURL := fmt.Sprintf(
		"%s/storage/v1/object/%s",
		SupabaseURL,
		bucket,
	)

	bodyBytes, err := json.Marshal(map[string][]string{
		"prefixes": {objectPath},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, deleteURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+SupabaseKey)
	req.Header.Set("apikey", SupabaseKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("❌ delete falhou (%d): %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("✅ Arquivo deletado do Supabase: %s | resposta: %s\n", objectPath, string(respBody))
	return nil
}

// slugify transforma texto em slug seguro
func slugify(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = strings.ReplaceAll(s, " ", "-")

	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	return reg.ReplaceAllString(s, "")
}

// ExtractObjectPath extrai o caminho relativo do objeto a partir da URL pública
func ExtractObjectPath(publicURL string, bucket string) (string, error) {
	marker := fmt.Sprintf("/storage/v1/object/public/%s/", bucket)
	idx := strings.Index(publicURL, marker)
	if idx == -1 {
		return "", fmt.Errorf("URL inválida, não foi possível extrair o caminho: %s", publicURL)
	}
	return publicURL[idx+len(marker):], nil
}