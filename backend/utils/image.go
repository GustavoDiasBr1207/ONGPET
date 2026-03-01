package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"strings"

	"github.com/disintegration/imaging"
)

/*
	OptimizedImage representa uma imagem já tratada
	pronta para upload (buffer + metadados)
*/
type OptimizedImage struct {
	Buffer      *bytes.Buffer // imagem final pronta para upload
	ContentType string        // image/jpeg ou image/png
	Extension   string        // .jpg ou .png
	Width       int           // largura final
	Height      int           // altura final
}

/*
	ResizeAndCompressImage

	- Decodifica a imagem enviada (jpeg, png, webp)
	- Redimensiona mantendo proporção
	- Comprime JPEG (qualidade configurável)
	- PNG mantém transparência
	- Remove EXIF automaticamente (image.Decode)
*/
func ResizeAndCompressImage(
	file *multipart.FileHeader,
	maxWidth int,    // ex: 1280
	jpegQuality int, // 1–100 (recomendado: 70–85)
) (*OptimizedImage, error) {

	// 🔹 Abre arquivo
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 🔹 Decodifica imagem (strip EXIF automaticamente)
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar imagem: %w", err)
	}

	// 🔹 Redimensiona mantendo proporção
	if img.Bounds().Dx() > maxWidth {
		img = imaging.Resize(img, maxWidth, 0, imaging.Lanczos)
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	buf := new(bytes.Buffer)

	switch strings.ToLower(format) {

	case "jpeg", "jpg":
		err = jpeg.Encode(buf, img, &jpeg.Options{
			Quality: jpegQuality,
		})
		if err != nil {
			return nil, err
		}
		return &OptimizedImage{
			Buffer:      buf,
			ContentType: "image/jpeg",
			Extension:   ".jpg",
			Width:       width,
			Height:      height,
		}, nil

	case "png":
		// PNG mantém transparência (sem perda)
		err = png.Encode(buf, img)
		if err != nil {
			return nil, err
		}
		return &OptimizedImage{
			Buffer:      buf,
			ContentType: "image/png",
			Extension:   ".png",
			Width:       width,
			Height:      height,
		}, nil

	case "webp":
		// 🔥 WebP → JPEG otimizado (melhor compatibilidade)
		err = jpeg.Encode(buf, img, &jpeg.Options{
			Quality: jpegQuality,
		})
		if err != nil {
			return nil, err
		}
		return &OptimizedImage{
			Buffer:      buf,
			ContentType: "image/jpeg",
			Extension:   ".jpg",
			Width:       width,
			Height:      height,
		}, nil

	default:
		return nil, fmt.Errorf("formato de imagem não suportado: %s", format)
	}
}