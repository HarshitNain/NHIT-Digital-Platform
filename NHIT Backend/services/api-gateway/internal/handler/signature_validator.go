package handler

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

// SignatureValidationConfig holds validation thresholds
type SignatureValidationConfig struct {
	MaxColors        int     // Maximum unique colors allowed
	MinWidth         int     // Minimum width in pixels
	MinHeight        int     // Minimum height in pixels
	MaxWidth         int     // Maximum width in pixels
	MaxHeight        int     // Maximum height in pixels
	MinAspectRatio   float64 // Width/Height ratio (signatures are horizontal)
	MaxFileSizeKB    int64   // Maximum file size in KB
	MinBgCoverage    float64 // Minimum % of background color
}

// DefaultSignatureConfig returns sensible defaults for signature validation
func DefaultSignatureConfig() SignatureValidationConfig {
	return SignatureValidationConfig{
		MaxColors:      300,  // Increased for JPEG compression + anti-aliasing
		MinWidth:       100,  // Minimum 100px wide
		MinHeight:      30,   // Minimum 30px tall
		MaxWidth:       2000, // Maximum 2000px wide
		MaxHeight:      1000, // Maximum 1000px tall
		MinAspectRatio: 1.0,  // Width should be at least equal to height (allows square signatures)
		MaxFileSizeKB:  300,  // 300KB max (signatures compress well)
		MinBgCoverage:  0.40, // Reduced to 40% for signatures with more ink coverage
	}
}

// ValidateSignatureImage validates that an uploaded file is a legitimate signature
func ValidateSignatureImage(file multipart.File, header *multipart.FileHeader, config SignatureValidationConfig) error {
	// 1. Read file bytes for type detection and image decoding
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Reset file pointer for potential reuse
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	// 2. Validate file type from actual bytes (not just extension)
	contentType := http.DetectContentType(fileBytes)
	log.Printf("🔍 Signature validation - File type detected: %s", contentType)
	if contentType != "image/jpeg" && contentType != "image/png" {
		return fmt.Errorf("invalid file type: only JPEG and PNG images are allowed for signatures")
	}

	// 3. Check file size
	fileSizeKB := int64(len(fileBytes)) / 1024
	log.Printf("🔍 Signature validation - File size: %dKB (max: %dKB)", fileSizeKB, config.MaxFileSizeKB)
	if fileSizeKB > config.MaxFileSizeKB {
		return fmt.Errorf("file size too large: %dKB (max: %dKB). Signatures should be small files", fileSizeKB, config.MaxFileSizeKB)
	}

	// 4. Decode image to analyze pixels
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// 5. Check dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width < config.MinWidth || height < config.MinHeight {
		log.Printf("❌ Signature validation failed - Image too small: %dx%d", width, height)
		return fmt.Errorf("image too small: %dx%d (minimum: %dx%d)", width, height, config.MinWidth, config.MinHeight)
	}

	if width > config.MaxWidth || height > config.MaxHeight {
		log.Printf("❌ Signature validation failed - Image too large: %dx%d", width, height)
		return fmt.Errorf("image too large: %dx%d (maximum: %dx%d)", width, height, config.MaxWidth, config.MaxHeight)
	}

	// 6. Check aspect ratio (signatures are typically horizontal)
	aspectRatio := float64(width) / float64(height)
	log.Printf("🔍 Signature validation - Dimensions: %dx%d, Aspect ratio: %.2f (min: %.2f)", width, height, aspectRatio, config.MinAspectRatio)
	if aspectRatio < config.MinAspectRatio {
		log.Printf("❌ Signature validation failed - Invalid aspect ratio")
		return fmt.Errorf("invalid aspect ratio: %.2f (signatures should be horizontal, minimum ratio: %.2f)", aspectRatio, config.MinAspectRatio)
	}

	// 7. Analyze color palette
	colorMap := make(map[uint32]int)
	totalPixels := width * height

	// Sample pixels to count colors (for performance, sample every Nth pixel for large images)
	sampleRate := 1
	if totalPixels > 100000 {
		sampleRate = 2 // Sample every 2nd pixel for images > 100k pixels
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			r, g, b, a := img.At(x, y).RGBA()
			// Convert to 8-bit and create color key
			colorKey := (uint32(r>>8) << 24) | (uint32(g>>8) << 16) | (uint32(b>>8) << 8) | uint32(a>>8)
			colorMap[colorKey]++
		}
	}

	uniqueColors := len(colorMap)
	log.Printf("🔍 Signature validation - Unique colors detected: %d (max: %d)", uniqueColors, config.MaxColors)
	if uniqueColors > config.MaxColors {
		log.Printf("❌ Signature validation failed - Too many colors")
		return fmt.Errorf("too many colors detected: %d (max: %d). This appears to be a photo or complex graphic, not a signature", uniqueColors, config.MaxColors)
	}

	// 8. Find dominant color (should be background - white or transparent)
	var dominantColor uint32
	var maxCount int
	for color, count := range colorMap {
		if count > maxCount {
			maxCount = count
			dominantColor = color
		}
	}

	// Extract dominant color components
	domR := uint8((dominantColor >> 24) & 0xFF)
	domG := uint8((dominantColor >> 16) & 0xFF)
	domB := uint8((dominantColor >> 8) & 0xFF)
	domA := uint8(dominantColor & 0xFF)

	// Check if dominant color is white-ish or transparent
	isWhitish := domR > 240 && domG > 240 && domB > 240
	isTransparent := domA < 50
	log.Printf("🔍 Signature validation - Dominant color: R=%d G=%d B=%d A=%d, Whitish: %v, Transparent: %v", domR, domG, domB, domA, isWhitish, isTransparent)
	
	if !isWhitish && !isTransparent {
		log.Printf("❌ Signature validation failed - Invalid background color")
		return fmt.Errorf("invalid background: signatures should have white or transparent backgrounds")
	}

	// 9. Check background coverage
	sampledPixels := (width / sampleRate) * (height / sampleRate)
	bgCoverage := float64(maxCount) / float64(sampledPixels)
	log.Printf("🔍 Signature validation - Background coverage: %.1f%% (min: %.1f%%)", bgCoverage*100, config.MinBgCoverage*100)
	
	if bgCoverage < config.MinBgCoverage {
		log.Printf("❌ Signature validation failed - Insufficient background coverage")
		return fmt.Errorf("insufficient background coverage: %.1f%% (minimum: %.1f%%). Signature should have clear background", bgCoverage*100, config.MinBgCoverage*100)
	}

	// All checks passed!
	log.Printf("✅ Signature validation PASSED - All checks successful")
	return nil
}
