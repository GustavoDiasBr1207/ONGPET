package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	SupabaseURL               string
	SupabaseKey               string
	SupabaseBucketPets        string
	SupabaseBucketOngs        string
	SupabaseBucketFormularios string
	SupabaseBucketBanners     string
)

// InitSupabase inicializa variáveis do Supabase Storage (REST API)
func InitSupabase() {
	SupabaseURL = strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	SupabaseKey = strings.TrimSpace(os.Getenv("SUPABASE_KEY")) // ← anon key

	SupabaseBucketPets        = strings.TrimSpace(os.Getenv("SUPABASE_BUCKET_PETS"))
	SupabaseBucketOngs        = strings.TrimSpace(os.Getenv("SUPABASE_BUCKET_ONGS"))
	SupabaseBucketFormularios = strings.TrimSpace(os.Getenv("SUPABASE_BUCKET_FORMULARIOS"))
	SupabaseBucketBanners     = strings.TrimSpace(os.Getenv("SUPABASE_BUCKET_BANNERS"))

	if SupabaseURL == "" || SupabaseKey == "" {
		panic("❌ SUPABASE_URL e SUPABASE_KEY são obrigatórios")
	}
	if SupabaseBucketPets == "" || SupabaseBucketOngs == "" || SupabaseBucketFormularios == "" || SupabaseBucketBanners == "" {
		panic("❌ SUPABASE_BUCKET_PETS, SUPABASE_BUCKET_ONGS, SUPABASE_BUCKET_FORMULARIOS e SUPABASE_BUCKET_BANNERS são obrigatórios")
	}

	fmt.Println("✅ Supabase Storage (REST) inicializado com sucesso")
	fmt.Printf("   bucket pets:        %s\n", SupabaseBucketPets)
	fmt.Printf("   bucket ongs:        %s\n", SupabaseBucketOngs)
	fmt.Printf("   bucket formularios: %s\n", SupabaseBucketFormularios)
	fmt.Printf("   bucket banners:     %s\n", SupabaseBucketBanners)
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


func DeleteFile(bucket string, objectPath string, authToken string) error {
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

	bearerToken := authToken
	if bearerToken == "" {
		bearerToken = SupabaseKey
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
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