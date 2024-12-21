#!/bin/bash

set -e

if [[ ! -d backend ]]; then
  echo "Must be run from root of repo."
  exit 1
fi

pushd backend
go test -v ./...
GOOS=linux GOARCH=amd64 go build
popd
pushd frontend
npm run build
popd

rsync -avz --progress -e "ssh -p 21098" backend/backend codertks@eot.coderealms.io:eot.coderealms.io/public_html/cgi-bin/api.cgi
rsync -avz -e "ssh -p 21098" frontend/dist/ codertks@eot.coderealms.io:eot.coderealms.io/public_html/
rsync -avz -e "ssh -p 21098" deployment/htaccess codertks@eot.coderealms.io:eot.coderealms.io/public_html/.htaccess

if [ -e deployment/config-prod.json ]; then
  rsync -avz -e "ssh -p 21098" deployment/config-prod.json codertks@eot.coderealms.io:eot.coderealms.io/backend/config-prod.json
fi
