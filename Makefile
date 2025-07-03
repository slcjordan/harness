.SECONDARY:

DEV_NAMESPACE?=harness-$(shell git rev-parse --abbrev-ref HEAD)
GRPCMAN_VERSION?=1.2.1
PORTAL_VERSION?=v1.0.50

.PHONY: docker-build-grpcman
docker-build-grpcman: .cache/${DEV_NAMESPACE}/artifacts/_grpcman_${GRPCMAN_VERSION}.AppImage.extracted
	docker build \
		--file docker/grpcman \
		--build-arg APP_DIR=$</squashfs-root \
		--tag ${DEV_NAMESPACE}-grpcman \
		.

.PHONY: docker-build-harness
docker-build-harness:
	DOCKER_BUILDKIT=1 docker build \
		--secret id=netrc,src=/home/jcrabtree/.netrc \
		--file docker/harness \
		--tag ${DEV_NAMESPACE}-harness \
		--build-arg PORTAL_VERSION=${PORTAL_VERSION} \
		.

.PHONY: docker-run-harness
docker-run-harness: docker-build-harness
	docker run \
		--rm \
		--interactive \
		--tty \
		--publish 1984:1984 \
		--volume ${PWD}/ui/public:/srv/harness \
		${DEV_NAMESPACE}-harness

.cache/${DEV_NAMESPACE}/artifacts/grpcman_%.AppImage:
	mkdir --parents $(@D)
	curl --output $@ https://tp-artifactory.vivint.com/artifactory/plainfiles/grpcman/grpcman_$*.AppImage

.cache/${DEV_NAMESPACE}/artifacts/_%.AppImage.extracted: .cache/${DEV_NAMESPACE}/artifacts/%.AppImage
	binwalk --extract $<
	unsquashfs $@/*.squashfs -dest $(<D)/squashfs-root
