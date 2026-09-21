# Development
run:
	hugo server --buildFuture

# Theme building
build-theme:
	cd themes/heimatverein-niederjosbach/bootstrap-sass && \
	bundle install --path vendor/bundle && \
	bundle exec compass compile

build: build-theme
	hugo build --buildFuture

# Gallery workflow
# See helpers/gallery-generator/README.md for the full explanation.
#
# Local config and credentials live in .env (gitignored). Start with:
#   cp .env.example .env
#
# Usage:
#   make gallery-generate            # uses GALLERY_SOURCE from .env
#   make gallery-generate SOURCE=/path/to/source/galleries
#   make gallery-push
#
# Config (override on the command line if needed):
ENV_FILE       ?= .env
SOURCE         ?=
GALLERY_TARGET ?= public/images/galerie
GALLERY_DATA   ?= data/galleries

# Sourced by the shell (not by make), so values may contain $, # etc. as
# long as they are quoted like in a normal shell script.
LOAD_ENV = if [ -f "$(ENV_FILE)" ]; then set -a; . "$(abspath $(ENV_FILE))"; set +a; fi

# Copy new images from the source folder, generate thumbnails, and write
# the JSON manifests to data/galleries/. Only touches what's missing,
# unless FORCE=1 is passed to also regenerate existing thumbnails.
gallery-generate:
	@$(LOAD_ENV); \
	src="$(SOURCE)"; src="$${src:-$$GALLERY_SOURCE}"; \
	if [ -z "$$src" ]; then \
		echo "Set GALLERY_SOURCE in $(ENV_FILE) or pass SOURCE=/path/to/source/galleries" >&2; exit 1; \
	fi; \
	cd helpers/gallery-generator && go run main.go \
		-source "$$src" \
		-target "$(CURDIR)/$(GALLERY_TARGET)" \
		-datadir "$(CURDIR)/$(GALLERY_DATA)" \
		-full $(if $(FORCE),-force)

# Upload the locally generated gallery images (public/images/galerie) to
# the webserver via SFTP. Credentials come from $(ENV_FILE); the password
# is handed to lftp via LFTP_PASSWORD so it doesn't show up in `ps`.
gallery-push:
	@$(LOAD_ENV); \
	: "$${FTP_HOST:?not set in $(ENV_FILE)}" "$${FTP_USER:?not set in $(ENV_FILE)}" "$${FTP_PASS:?not set in $(ENV_FILE)}"; \
	LFTP_PASSWORD="$$FTP_PASS" lftp --env-password -u "$$FTP_USER" "sftp://$$FTP_HOST" -e \
		"set sftp:auto-confirm yes; mirror -R --parallel=4 $(GALLERY_TARGET) $${FTP_REMOTE_DIR:-images/galerie}; bye"

.PHONY: run build-theme build gallery-generate gallery-push
