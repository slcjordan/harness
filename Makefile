.SECONDARY:

DEV_NAMESPACE?=harness-$(shell git rev-parse --abbrev-ref HEAD)
GRPCMAN_VERSION?=1.2.1

.PHONY: download-grpcman
download-grpcman: .cache/${DEV_NAMESPACE}/artifacts/_grpcman_${GRPCMAN_VERSION}.AppImage.extracted
	docker build \
		--file docker/grpcman \
		--build-arg APP_DIR=$</squashfs-root \
		--tag ${DEV_NAMESPACE}-grpcman \
		.

.cache/${DEV_NAMESPACE}/artifacts/grpcman_%.AppImage:
	mkdir --parents $(@D)
	curl --output $@ https://tp-artifactory.vivint.com/artifactory/plainfiles/grpcman/grpcman_$*.AppImage

.cache/${DEV_NAMESPACE}/artifacts/_%.AppImage.extracted: .cache/${DEV_NAMESPACE}/artifacts/%.AppImage
	binwalk --extract $<
	unsquashfs $@/*.squashfs -dest $(<D)/squashfs-root
