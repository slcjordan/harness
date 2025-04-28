
DEV_NAMESPACE?=harness-$(shell git rev-parse --abbrev-ref HEAD)

.PHONY: download-grpcman
download-grpcman: .cache/${DEV_NAMESPACE}/downloads/grpcman_1.2.1.AppImage:

.cache/${DEV_NAMESPACE}/downloads/grpcman_1.2.1.AppImage:
	mkdir -p $(@D)
	curl -O https://tp-artifactory.vivint.com/artifactory/plainfiles/grpcman/grpcman_1.2.1.AppImage
