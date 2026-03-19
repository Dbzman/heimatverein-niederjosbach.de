package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/disintegration/imaging"
)

type GalleryData struct {
	Year   string       `json:"year"`
	Event  string       `json:"event"`
	Path   string       `json:"path"`
	Images []ImageEntry `json:"images"`
	Count  int          `json:"count"`
}

type ImageEntry struct {
	Original  string `json:"original"`
	Thumbnail string `json:"thumbnail"`
}

const (
	thumbnailSize   = 400
	thumbnailSuffix = "_thumb"
)

func main() {
	// Command line flags
	sourceDir := flag.String("source", "", "Source gallery directory (required only for -copy)")
	targetDir := flag.String("target", "", "Target gallery directory (in repo, default: public/images/galerie)")
	dataDir := flag.String("datadir", "", "Data output directory (default: data/galleries)")
	copyImages := flag.Bool("copy", false, "Copy images from source to target directory")
	generateThumbs := flag.Bool("thumbnails", false, "Generate thumbnails for images in target directory")
	generateData := flag.Bool("data", false, "Generate JSON data files from target directory")
	full := flag.Bool("full", false, "Enable all operations (copy, thumbnails, and data)")
	force := flag.Bool("force", false, "Force regeneration of thumbnails even if they exist")
	flag.Parse()

	// If full is set, enable all operations
	if *full {
		*copyImages = true
		*generateThumbs = true
		*generateData = true
	}

	// At least one operation must be specified
	if !*copyImages && !*generateThumbs && !*generateData {
		fmt.Println("Error: At least one operation must be specified: -copy, -thumbnails, -data, or -full")
		os.Exit(1)
	}

	// Validate source directory only if copying
	if *copyImages && *sourceDir == "" {
		fmt.Println("Error: -source flag is required when using -copy")
		fmt.Println("Usage: go run main.go -source /path/to/source/galleries -copy [-thumbnails] [-data]")
		os.Exit(1)
	}

	// Set defaults for target and data directories
	projectRoot := filepath.Join("..", "..")
	if *targetDir == "" {
		*targetDir = filepath.Join(projectRoot, "public", "images", "galerie")
	}
	if *dataDir == "" {
		*dataDir = filepath.Join(projectRoot, "data", "galleries")
	}

	if *copyImages {
		fmt.Printf("Source directory: %s\n", *sourceDir)
	}
	fmt.Printf("Target directory: %s\n", *targetDir)
	fmt.Printf("Data directory: %s\n", *dataDir)
	fmt.Printf("Copy images: %v\n", *copyImages)
	fmt.Printf("Generate thumbnails: %v\n", *generateThumbs)
	fmt.Printf("Generate data files: %v\n", *generateData)
	if *force {
		fmt.Printf("Force mode: enabled\n")
	}
	fmt.Println()

	// Process galleries
	if err := processGalleries(*sourceDir, *targetDir, *dataDir, *copyImages, *generateThumbs, *generateData, *force); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nGallery processing complete!")
}

func processGalleries(sourceDir, targetDir, dataDir string, copyImages, generateThumbs, generateData, force bool) error {
	// Determine base directory for reading structure
	baseDir := targetDir
	if copyImages {
		baseDir = sourceDir
		// Check if source directory exists
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			return fmt.Errorf("source directory does not exist: %s", sourceDir)
		}
	}

	// Check if base directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", baseDir)
	}

	// Read year directories
	yearDirs, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	totalImages := 0
	totalCopied := 0
	totalThumbnails := 0

	for _, yearDir := range yearDirs {
		if !yearDir.IsDir() || strings.HasPrefix(yearDir.Name(), ".") {
			continue
		}

		year := yearDir.Name()
		fmt.Printf("\nProcessing year: %s\n", year)

		// Read event directories
		yearPath := filepath.Join(baseDir, year)
		eventDirs, err := os.ReadDir(yearPath)
		if err != nil {
			fmt.Printf("Warning: could not read %s: %v\n", yearPath, err)
			continue
		}

		for _, eventDir := range eventDirs {
			if !eventDir.IsDir() || strings.HasPrefix(eventDir.Name(), ".") {
				continue
			}

			event := eventDir.Name()
			fmt.Printf("  Processing event: %s\n", event)

			var sourceEventPath, targetEventPath string
			if copyImages {
				sourceEventPath = filepath.Join(sourceDir, year, event)
				targetEventPath = filepath.Join(targetDir, year, event)
			} else {
				targetEventPath = filepath.Join(targetDir, year, event)
			}

			// Create target directory only if we're copying images or generating thumbnails
			if copyImages || generateThumbs {
				if err := os.MkdirAll(targetEventPath, 0755); err != nil {
					return fmt.Errorf("failed to create target directory %s: %w", targetEventPath, err)
				}
			}

			// Process images in event directory
			images, copiedCount, thumbCount, err := processEventImages(sourceEventPath, targetEventPath, copyImages, generateThumbs, force)
			if err != nil {
				fmt.Printf("    Warning: error processing event %s/%s: %v\n", year, event, err)
				continue
			}

			if len(images) > 0 {
				totalImages += len(images)
				totalCopied += copiedCount
				totalThumbnails += thumbCount

				// Build status message
				var statusParts []string
				statusParts = append(statusParts, fmt.Sprintf("%d images", len(images)))
				if copyImages {
					statusParts = append(statusParts, fmt.Sprintf("copied %d", copiedCount))
				}
				if generateThumbs {
					statusParts = append(statusParts, fmt.Sprintf("generated %d thumbnails", thumbCount))
				}
				fmt.Printf("    Processed %s\n", strings.Join(statusParts, ", "))

				// Write JSON manifest
				if generateData {
					if err := writeGalleryJSON(year, event, images, dataDir); err != nil {
						return fmt.Errorf("failed to write JSON for %s/%s: %w", year, event, err)
					}
				}
			}
		}
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  Total images: %d\n", totalImages)
	if copyImages {
		fmt.Printf("  Images copied: %d\n", totalCopied)
	}
	if generateThumbs {
		fmt.Printf("  Thumbnails generated: %d\n", totalThumbnails)
	}

	return nil
}

func processEventImages(sourceEventPath, targetEventPath string, copyImages, generateThumbs, force bool) ([]ImageEntry, int, int, error) {
	// Determine which directory to read from
	readDir := targetEventPath
	if copyImages {
		readDir = sourceEventPath
	}

	files, err := os.ReadDir(readDir)
	if err != nil {
		return nil, 0, 0, err
	}

	var images []ImageEntry
	imagesCopied := 0
	thumbnailsGenerated := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		// Skip thumbnails when reading from target
		if !copyImages && strings.Contains(file.Name(), thumbnailSuffix) {
			continue
		}

		targetPath := filepath.Join(targetEventPath, file.Name())

		// Copy original image if it doesn't exist or is different
		if copyImages {
			sourcePath := filepath.Join(sourceEventPath, file.Name())
			if needsCopy(sourcePath, targetPath) {
				if err := copyFile(sourcePath, targetPath); err != nil {
					fmt.Printf("      Warning: failed to copy %s: %v\n", file.Name(), err)
					continue
				}
				imagesCopied++
			}
		}

		// Generate thumbnail
		thumbName := generateThumbnailName(file.Name())
		thumbPath := filepath.Join(targetEventPath, thumbName)

		if generateThumbs {
			shouldGenerate := force || needsThumbnail(targetPath, thumbPath)
			if shouldGenerate {
				if err := generateThumbnail(targetPath, thumbPath); err != nil {
					fmt.Printf("      Warning: failed to generate thumbnail for %s: %v\n", file.Name(), err)
				} else {
					thumbnailsGenerated++
				}
			}
		}

		images = append(images, ImageEntry{
			Original:  file.Name(),
			Thumbnail: thumbName,
		})
	}

	sort.Slice(images, func(i, j int) bool {
		return images[i].Original < images[j].Original
	})

	return images, imagesCopied, thumbnailsGenerated, nil
}

func needsCopy(sourcePath, targetPath string) bool {
	targetInfo, err := os.Stat(targetPath)
	if os.IsNotExist(err) {
		return true
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return true
	}

	// Copy if sizes differ
	return sourceInfo.Size() != targetInfo.Size()
}

func needsThumbnail(originalPath, thumbPath string) bool {
	_, err := os.Stat(thumbPath)
	return os.IsNotExist(err)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func generateThumbnailName(originalName string) string {
	ext := filepath.Ext(originalName)
	nameWithoutExt := strings.TrimSuffix(originalName, ext)
	return nameWithoutExt + thumbnailSuffix + ext
}

func generateThumbnail(originalPath, thumbPath string) error {
	// Open and decode image
	img, err := imaging.Open(originalPath)
	if err != nil {
		return err
	}

	// Crop to square from center and resize to thumbnailSize
	thumb := imaging.Fill(img, thumbnailSize, thumbnailSize, imaging.Center, imaging.Lanczos)

	// Save thumbnail as JPEG with quality 85
	return imaging.Save(thumb, thumbPath, imaging.JPEGQuality(85))
}

func writeGalleryJSON(year, event string, images []ImageEntry, dataDir string) error {
	yearDir := filepath.Join(dataDir, year)
	if err := os.MkdirAll(yearDir, 0755); err != nil {
		return err
	}

	galleryData := GalleryData{
		Year:   year,
		Event:  event,
		Path:   fmt.Sprintf("%s/%s", year, event),
		Images: images,
		Count:  len(images),
	}

	jsonData, err := json.MarshalIndent(galleryData, "", "  ")
	if err != nil {
		return err
	}

	outputFile := filepath.Join(yearDir, event+".json")
	return os.WriteFile(outputFile, jsonData, 0644)
}
