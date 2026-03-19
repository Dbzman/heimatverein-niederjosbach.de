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

.PHONY: run build-theme build
