NOWISH := $(shell date +%s)

.PHONY: build
build:
	rm -rf .build
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o transmission-jobs .
	install -d -m 755 .build/var/lib/transmission-jobs/
	install -D -m 755 transmission-jobs .build/bin/transmission-jobs && rm transmission-jobs
	install -D -m 644 .dist/transmission-jobs.service .build/usr/lib/systemd/system/transmission-jobs.service
	install -D -m 644 .dist/transmission-jobs.timer .build/usr/lib/systemd/system/transmission-jobs.timer
	install -D -m 600 transmission-jobs.default.yml .build/etc/transmission-jobs.yml

.PHONY: deb
deb: build
	rm -f *.deb
	fpm -s dir \
		-t deb \
		-C .build \
		--name transmission-jobs \
		--version '$(NOWISH)' \
		--maintainer 'Mark Ignacio <mark@ignacio.io>' \
		--before-install .dist/before-install.sh \
		--after-install .dist/after-install.sh \
		--config-files etc/transmission-jobs.yml \
		.