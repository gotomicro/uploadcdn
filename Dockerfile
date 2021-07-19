ARG APP=uploadcdn
FROM alpine
ARG APP
ENV APP=${APP}
ENV WORKDIR=/data
COPY bin ${WORKDIR}
WORKDIR ${WORKDIR}
CMD ["sh", "-c", "./${APP}"]
