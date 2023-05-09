#!/usr/bin/env bash

./build/build.sh

buildah from --name puzzlesettingsserver-working-container scratch
buildah copy puzzlesettingsserver-working-container $HOME/go/bin/puzzlesettingsserver /bin/puzzlesettingsserver
buildah config --env SERVICE_PORT=50051 puzzlesettingsserver-working-container
buildah config --port 50051 puzzlesettingsserver-working-container
buildah config --entrypoint '["/bin/puzzlesettingsserver"]' puzzlesettingsserver-working-container
buildah commit puzzlesettingsserver-working-container puzzlesettingsserver
buildah rm puzzlesettingsserver-working-container

buildah push puzzlesettingsserver docker-daemon:puzzlesettingsserver:latest
