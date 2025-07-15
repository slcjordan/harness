.SECONDARY:

DEV_NAMESPACE?=harness-$(shell git rev-parse --abbrev-ref HEAD)
GRPCMAN_VERSION?=1.2.1
PORTAL_VERSION?=v1.0.50
HUGO_VERSION?=reg-git-non-root-0.136.5

.PHONY: docker-build-grpcman
docker-build-grpcman: .cache/${DEV_NAMESPACE}/artifacts/_grpcman_${GRPCMAN_VERSION}.AppImage.extracted
	docker build \
		--file docker/grpcman \
		--build-arg APP_DIR=$</squashfs-root \
		--tag ${DEV_NAMESPACE}-grpcman \
		.

.PHONY: hugo-build
hugo-build:
	docker run \
		--rm \
		--interactive \
		--tty \
		--volume ${PWD}/ui/:/ui \
		hugomods/hugo:${HUGO_VERSION} \
			hugo --source /ui

.PHONY: docker-build-harness
docker-build-harness:
	DOCKER_BUILDKIT=1 docker build \
		--secret id=netrc,src=/home/jcrabtree/.netrc \
		--file docker/harness \
		--tag ${DEV_NAMESPACE}-harness \
		--build-arg PORTAL_VERSION=${PORTAL_VERSION} \
		.

# TODO random port
.PHONY: docker-run-harness
docker-run-harness: docker-build-harness
	docker run \
		--rm \
		--interactive \
		--tty \
		--publish 1984:1984 \
		--volume ${PWD}/ui/public:/srv/harness \
		${DEV_NAMESPACE}-harness

.PHONY: docker-build-harness-buildkit
docker-build-harness-buildkit: hugo-build
	DOCKER_BUILDKIT=1 docker build \
		--secret id=netrc,src=/home/jcrabtree/.netrc \
		--file docker/harness-buildkit \
		--tag ${DEV_NAMESPACE}-harness-buildkit \
		--build-arg PORTAL_VERSION=${PORTAL_VERSION} \
		.

.PHONY: docker-run-harness-buildkit
docker-run-harness-buildkit: docker-build-harness-buildkit
	docker run \
		--rm \
		--interactive \
		--tty \
		--publish 1984:1984 \
		--volume ${PWD}/ui/public:/srv/harness \
		${DEV_NAMESPACE}-harness-buildkit

.PHONY: docker-build-grpcui
docker-build-grpcui:
	DOCKER_BUILDKIT=1 docker build \
		--file docker/grpcui \
		--tag ${DEV_NAMESPACE}-grpcui \
		.

.PHONY: docker-run-grpcui
docker-run-grpcui: docker-build-grpcui
	docker run \
		--rm \
		--interactive \
		--tty \
		--volume /home/jcrabtree/Developer/src/:/home/jcrabtree/Developer/src/ \
		--network host \
		${DEV_NAMESPACE}-grpcui \
			--import-path /include/proto \
			--import-path /home/jcrabtree/Developer/src/gitlab.com/vivint/horizontals/platform/device/ \
			--proto capabilities/actions/actions.proto \
			-plaintext \
			device:8856

.cache/${DEV_NAMESPACE}/artifacts/grpcman_%.AppImage:
	mkdir --parents $(@D)
	curl --output $@ https://tp-artifactory.vivint.com/artifactory/plainfiles/grpcman/grpcman_$*.AppImage

.cache/${DEV_NAMESPACE}/artifacts/_%.AppImage.extracted: .cache/${DEV_NAMESPACE}/artifacts/%.AppImage
	binwalk --extract $<
	unsquashfs $@/*.squashfs -dest $(<D)/squashfs-root
