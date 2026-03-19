# Gallery Generator

Processes gallery images and generates JSON manifests for Hugo consumption.

## Features

- **Copies images** from source folder (outside repo) to target folder (in repo)
- **Generates thumbnails** (400px width) for all images
- **Creates JSON manifests** for Hugo to consume
- **Incremental processing** - only copies/generates what's missing

## Usage

```bash
cd helpers/gallery-generator
go run main.go -source /path/to/source/galleries
```

### Command Line Options

- `-source` (required): Source gallery directory with structure `<year>/<event>/*.jpg`
- `-target` (optional): Target directory (default: `../../public/images/galerie`)
- `-data` (optional): Data output directory (default: `../../data/galleries`)

## Directory Structure

### Source (Outside Repo)
```
source/
├── 2024/
│   ├── dickworzschnitzen_2024/
│   │   ├── image-01.jpg
│   │   ├── image-02.jpg
│   │   └── ...
│   └── sommerfest_2024/
│       └── ...
└── 2023/
    └── ...
```

### Target (Hugo Public Folder)
```
public/images/galerie/
├── 2024/
│   ├── dickworzschnitzen_2024/
│   │   ├── image-01.jpg
│   │   ├── image-01_thumb.jpg
│   │   ├── image-02.jpg
│   │   ├── image-02_thumb.jpg
│   │   └── ...
│   └── ...
└── ...
```

**Note:** Images go to `public/` folder which:
- Is NOT committed to git
- Is used by Hugo for rendering
- Gets deployed to Hetzner shared host

### Data Output
```
data/galleries/
├── 2024/
│   ├── dickworzschnitzen_2024.json
│   └── sommerfest_2024.json
└── ...
```

## JSON Output Format

```json
{
  "year": "2024",
  "event": "dickworzschnitzen_2024",
  "path": "2024/dickworzschnitzen_2024",
  "images": [
    {
      "original": "image-01.jpg",
      "thumbnail": "image-01_thumb.jpg"
    }
  ],
  "count": 1
}
```

## Workflow

1. **Run the generator** when new galleries are added to the source folder
2. **Commit the JSON files** (only `data/galleries/` - images stay out of git)
3. **Deploy to Hetzner**: Copy `public/images/galerie/` to server
4. **Hugo builds** use the JSON files to render galleries (fast, no image processing)

## Configuration

- Thumbnail width: 400px (configurable in `main.go`)
- Thumbnail suffix: `_thumb` (configurable in `main.go`)
- JPEG quality: 85 (configurable in `main.go`)

## Dependencies

- `github.com/disintegration/imaging` - Modern image processing library with high-quality resizing

Install with:
```bash
go mod download
```
