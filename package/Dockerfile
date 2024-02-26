#
# build the target application
#
FROM public.ecr.aws/docker/library/golang:1.21.7-alpine3.19 as builder
RUN apk update && apk upgrade && apk add --no-cache make npm gcc sqlite-dev musl-dev

# see: https://stackoverflow.com/questions/69692842/error-message-error0308010cdigital-envelope-routinesunsupported
ENV NODE_OPTIONS --openssl-legacy-provider

WORKDIR /build
COPY go.mod go.sum Makefile ./
COPY backend ./backend
COPY frontend ./frontend
COPY data ./data
RUN make linux-full

#
# build the target container
#
FROM public.ecr.aws/docker/library/alpine:3.19

# update the packages
RUN apk update && apk upgrade && apk add bash tzdata ca-certificates curl && rm -fr /var/cache/apk/*

# Create the run user and group
RUN addgroup --gid 18570 sse && adduser --uid 1984 docker -G sse -D

# set the timezone appropriatly
ENV TZ=UTC
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Specify home
ENV APP_HOME /libra3
WORKDIR $APP_HOME

# Create necessary directories
RUN mkdir -p $APP_HOME/scripts $APP_HOME/bin
RUN chown -R docker $APP_HOME && chgrp -R sse $APP_HOME

# port and run command
EXPOSE 8080
CMD scripts/entry.sh

# Move in necessary assets
COPY package/data/container_bash_profile /home/docker/.profile
COPY package/scripts/entry.sh $APP_HOME/scripts/entry.sh
COPY --from=builder /build/bin/libra3.linux $APP_HOME/bin/libra3
COPY --from=builder /build/bin/public $APP_HOME/bin/public
COPY --from=builder /build/bin/data $APP_HOME/bin/data

# Ensure permissions are correct
RUN chown docker:sse /home/docker/.profile $APP_HOME/scripts/entry.sh $APP_HOME/bin/libra3 && chmod 755 /home/docker/.profile $APP_HOME/scripts/entry.sh $APP_HOME/bin/libra3

# Add the build tag
ARG BUILD_TAG
RUN test -n "$BUILD_TAG" && touch $APP_HOME/buildtag.build-$BUILD_TAG || touch $APP_HOME/buildtag.build-0

# Specify the user
USER docker

#
# end of file
#
