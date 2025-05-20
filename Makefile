run:
	hugo server --buildFuture

build-theme:
	cd themes/heimatverein-niederjosbach/bootstrap-sass && \
	bundle install --path vendor/bundle && \
	bundle exec compass compile

build: build-theme
	hugo build

build-all: build-theme build
