# Gallery Generator

Processes gallery images and generates JSON manifests for Hugo to consume.
Images live **outside of git** and are pushed to the webserver separately
via SFTP; only the generated JSON manifests and the Hugo content pages are
committed to the repo.

## How a gallery ends up on the site

1. Photos are collected locally in a source folder that is **not** part of
   this repo (e.g. `~/Pictures/heimatverein-galerien/`), organized as
   `<year>/<event>/*.jpg`.
2. `gallery-generator` copies those images into `public/images/galerie/`,
   generates a 400px thumbnail for each one, and writes a JSON manifest per
   event to `data/galleries/<year>/<event>.json`.
3. The JSON manifests are committed to git. `public/images/galerie/` is
   **not** committed — it's gitignored, since it's just processed images.
4. A Hugo content page under `content/galerie/<year>/<slug>.md` is created
   (or already exists) with `galerie_pfad` pointing at `<year>/<event>` —
   this is what makes the gallery show up on `/galerie` and renders the
   images/thumbnails referenced in the JSON manifest.
5. The images themselves are uploaded straight to the webserver via SFTP
   (`make gallery-push`), independent of the normal Hugo build/deploy done
   by GitHub Actions on push to `master`. The CI build only ships HTML/CSS/JS
   from `public/`, not the gallery images — those need to already exist on
   the server, which is why they're pushed separately, once, from your
   machine.

All of the above is wrapped in top-level `Makefile` targets — see
[Quickstart](#quickstart) below.

## Quickstart: adding a new gallery

```bash
# 0. One-time setup: create your local config (gitignored) and fill it in
cp .env.example .env

# 1. Drop the photos in <year>/<event>/*.jpg under your local source folder,
#    e.g. ~/Pictures/heimatverein-galerien/2026/dickworzschnitzen_2026/

# 2. Generate: copies images + thumbnails into public/images/galerie/
#    and writes data/galleries/2026/dickworzschnitzen_2026.json
make gallery-generate

# 3. Create content/galerie/2026/dickworzschnitzen-2026.md (see below)

# 4. Preview locally
make run

# 5. Push the images to the webserver
make gallery-push

# 6. Commit + push the JSON manifest and the new content page
git add data/galleries/2026/dickworzschnitzen_2026.json \
        content/galerie/2026/dickworzschnitzen-2026.md
git commit -m "Add Dickworzschnitzen 2026 gallery"
git push
```

Step 6 triggers the normal CI build/deploy of the site itself; the images
were already placed on the server in step 5.

### The content page

Every gallery needs a Hugo page at `content/galerie/<year>/<slug>.md`
(filename can use dashes; `galerie_pfad` must match the `<year>/<event>`
folder name used in step 1/2, which conventionally uses underscores):

```toml
+++
date = "2026-09-01T12:00:00+02:00"
title = "Dickworzschnitzen 2026"
galerie_pfad = '2026/dickworzschnitzen_2026'
galerie_bild = 'IMG_0001.jpg'
+++
```

- `galerie_pfad`: `<year>/<event>` — must match the manifest path.
- `galerie_bild`: filename (in `public/images/galerie/<galerie_pfad>/`) used
  as the cover image on the `/galerie` overview page.

To link a news post (`content/aktuelles/...`) to the gallery, add
`galerie = '2026/dickworzschnitzen_2026'` to that post's front matter — it
will show a teaser image and a "Zur Galerie" button pointing at the gallery
page (see `themes/heimatverein-niederjosbach/layouts/aktuelles/single.html`).

## `make gallery-generate`

Wraps `go run main.go -source ... -full` (see [Command Line
Options](#command-line-options) below). Copies new/changed images, generates
missing thumbnails, and writes JSON manifests — safe to re-run, it only
touches what's missing. The source folder is taken from `GALLERY_SOURCE` in
`.env`, or from `SOURCE=` on the command line (which wins).

```bash
make gallery-generate
make gallery-generate FORCE=1                                 # regenerate all thumbnails
make gallery-generate SOURCE=/path/to/other/source/galleries  # one-off override
```

## `make gallery-push`

Uploads `public/images/galerie/` to the webserver (`FTP_REMOTE_DIR`, default
`images/galerie/`) via SFTP using `lftp`, mirroring only new/changed files.
Host, username and password are read from `FTP_HOST`, `FTP_USER` and
`FTP_PASS` in `.env`.

## Local config (`.env`)

All local settings and credentials for the Makefile live in `.env` at the
repo root. It is gitignored; `.env.example` is the committed template:

```bash
cp .env.example .env
```

| Variable         | Used by            | Description                                        |
|------------------|--------------------|----------------------------------------------------|
| `GALLERY_SOURCE` | `gallery-generate` | Local photo folder (`<year>/<event>/*.jpg`)        |
| `FTP_HOST`       | `gallery-push`     | SFTP host of the webserver                         |
| `FTP_USER`       | `gallery-push`     | SFTP username                                      |
| `FTP_PASS`       | `gallery-push`     | SFTP password                                      |
| `FTP_REMOTE_DIR` | `gallery-push`     | Target dir on the server (default `images/galerie`) |

The file is sourced by the shell, so quote values containing spaces, `$`,
`#` etc. with single quotes (e.g. `FTP_PASS='my$ecret#pw'`). To use a
different file, pass `ENV_FILE=path/to/file`.

## Manual usage

```bash
cd helpers/gallery-generator
go run main.go -source /path/to/source/galleries -full
```

### Command Line Options

- `-source`: Source gallery directory with structure `<year>/<event>/*.jpg`
  (required when using `-copy`)
- `-target` (optional): Target directory (default: `../../public/images/galerie`)
- `-datadir` (optional): Data output directory (default: `../../data/galleries`)
- `-copy`: Copy images from source to target directory
- `-thumbnails`: Generate thumbnails for images in the target directory
- `-data`: Generate JSON manifests from the target directory
- `-full`: Shorthand for `-copy -thumbnails -data`
- `-force`: Regenerate thumbnails even if they already exist

At least one of `-copy`, `-thumbnails`, `-data`, `-full` must be given.

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
- Gets uploaded to the Hetzner shared host separately, via `make gallery-push`

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

## Configuration

- Thumbnail width: 400px (configurable in `main.go`)
- Thumbnail suffix: `_thumb` (configurable in `main.go`)
- JPEG quality: 85 (configurable in `main.go`)

## Dependencies

- `github.com/disintegration/imaging` - image resizing
- `github.com/disintegration/imageorient` - EXIF-orientation-aware decoding
- [`lftp`](https://lftp.yar.ru/) for `make gallery-push`

Install Go deps with:
```bash
go mod download
```
