DOCKER_DIR=${CURDIR}
DOCKER_YML=${DOCKER_DIR}/docker-compose.yml
ENV_NAME="stage"

.PHONY: compose-up
compose-up:
	docker-compose -p ${ENV_NAME} -f ${DOCKER_YML} up -d

.PHONY: compose-down
compose-down:
	docker-compose -p ${ENV_NAME} -f ${DOCKER_YML} stop

.PHONY: compose-rm
compose-rm:
	docker-compose -p ${ENV_NAME} -f ${DOCKER_YML} rm -fvs

.PHONY: compose-rs
compose-rs:
	make compose-rm && \
 	make compose-up

