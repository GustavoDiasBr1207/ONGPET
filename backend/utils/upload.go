// utils/upload.go
package utils

import (
    "fmt"
    "log"
    "mime/multipart"
    "os"

    "github.com/nedpals/supabase-go"
)

// SupabaseClient é o cliente global do Supabase
var SupabaseClient *supabase.Client
var SupabaseBucketName string

// InitSupabase inicializa o cliente do Supabase
func InitSupabase() {
    url := os.Getenv("SUPABASE_URL")
    key := os.Getenv("SUPABASE_KEY")
    SupabaseBucketName = os.Getenv("SUPABASE_BUCKET") // exemplo: "pets"

    if url == "" || key == "" || SupabaseBucketName == "" {
        log.Fatal("Variáveis SUPABASE_URL, SUPABASE_KEY e SUPABASE_BUCKET devem estar setadas")
    }

    SupabaseClient = supabase.CreateClient(url, key)
}

// IsValidImage verifica se o arquivo é uma imagem válida (jpg, jpeg, png, gif)
func IsValidImage(file *multipart.FileHeader) bool {
    switch file.Header.Get("Content-Type") {
    case "image/jpeg", "image/png", "image/gif":
        return true
    default:
        return false
    }
}

// UploadFile envia um arquivo para o bucket e retorna a URL pública
func UploadFile(file *multipart.FileHeader, petID string) (string, error) {
    src, err := file.Open()
    if err != nil {
        return "", err
    }
    defer src.Close()

    objectPath := fmt.Sprintf("%s/%s", petID, file.Filename)

    // Faz upload usando nedpals/supabase-go
    SupabaseClient.Storage.From(SupabaseBucketName).Upload(objectPath, src, &supabase.FileUploadOptions{})

    // Retorna URL pública
    urlResp := SupabaseClient.Storage.From(SupabaseBucketName).GetPublicUrl(objectPath)
    if urlResp.SignedUrl == "" {
        return "", fmt.Errorf("não foi possível gerar URL pública para: %s", objectPath)
    }

    return urlResp.SignedUrl, nil
}

// UploadMultipleFiles envia múltiplos arquivos e retorna as URLs públicas
func UploadMultipleFiles(files []*multipart.FileHeader, petID string) ([]string, error) {
    var urls []string
    for _, file := range files {
        url, err := UploadFile(file, petID)
        if err != nil {
            return nil, err
        }
        urls = append(urls, url)
    }
    return urls, nil
}

// DeleteFile remove um arquivo do bucket
func DeleteFile(objectPath string) error {
    SupabaseClient.Storage.From(SupabaseBucketName).Remove([]string{objectPath})
    return nil
}

// DeleteFiles remove múltiplos arquivos do bucket
func DeleteFiles(objectPaths []string) error {
    SupabaseClient.Storage.From(SupabaseBucketName).Remove(objectPaths)
    return nil
}
