ARG CBOS_IMAGE_NAME
ARG CBOS_IMAGE_REPO
ARG CBOS_IMAGE_TAG

FROM ${CBOS_IMAGE_REPO}/${CBOS_IMAGE_NAME}:${CBOS_IMAGE_TAG}

COPY ./build/go-binary/eric-odp-cron-operator /usr/bin/eric-odp-cron-operator

ARG ERIC_ODP_CRON_OPERATOR_UID=261376
ARG ERIC_ODP_CRON_OPERATOR_GID=261376

ARG CBO_REPO=arm.rnd.ki.sw.ericsson.se/artifactory/proj-ldc-repo-rpm-local/common_base_os/sles/

ARG ARM_TOKEN
ARG STDOUT_URL=https://arm.seli.gic.ericsson.se/artifactory/proj-adp-log-release/com/ericsson/bss/adp/log/stdout-redirect/
ARG STDOUT_VERSION="1.36.0"
ARG CBOS_IMAGE_TAG

ARG GIT_COMMIT=""

RUN zypper ar -C -G -f https://${CBO_REPO}${CBOS_IMAGE_TAG}/?ssl_verify=no COMMON_BASE_OS_SLES_REPO && \
    zypper in -l -y \
    catatonit \
    curl

# RUN curl -fsSL -o /tmp/stdout-redirect.tar -H "Authorization:Bearer${ARM_TOKEN}" "${STDOUT_URL}/${STDOUT_VERSION}"/eric-log-libstdout-redirect-golang-cxa30176-"${STDOUT_VERSION}".x86_64.tar \
# && tar-C /-xf/tmp/stdout-redirect.tar \
# && rm-f /tmp/stdout-redirect.tar

# Workaround until stdout tar file is downloaded from above curl command.
COPY image_content/eric-log-libstdout-redirect-golang-cxa30176-1.37.0.x86_64.tar /tmp/
RUN tar -xvf /tmp/eric-log-libstdout-redirect-golang-cxa30176-1.37.0.x86_64.tar \
&& rm -f /tmp/eric-log-libstdout-redirect-golang-cxa30176-1.37.0.x86_64.tar

RUN echo "${ERIC_ODP_CRON_OPERATOR_UID}:x:${ERIC_ODP_CRON_OPERATOR_UID}:${ERIC_ODP_CRON_OPERATOR_GID}:eric-odp-cron-operator-user:/:/bin/bash" >> /etc/passwd && \
    cat /etc/passwd && \
    sed -i "s|root:/bin/bash|root:/bin/false|g" /etc/passwd && \
    chmod -R g=u /usr/bin/eric-odp-cron-operator && \
    chown -h ${ERIC_ODP_CRON_OPERATOR_UID}:0 /usr/bin/eric-odp-cron-operator

USER $ERIC_ODP_CRON_OPERATOR_UID:$ERIC_ODP_CRON_OPERATOR_GID

CMD ["catatonit", "--", "/stdout-redirect", "-redirect", "all", "-logfile", "/logs/cronoperator.log", "-size", "10", "-rotate", "3", "-run", "/usr/bin/eric-odp-cron-operator"]
