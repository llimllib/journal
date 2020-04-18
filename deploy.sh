#!/usr/bin/env bash
set -euxo pipefail

# set the linux build to be transferred
mv journal journal-local
cp journal-linux journal
trap "mv journal-local journal" EXIT

# it's dumb to have the whole server stopped while rsync'ing, but it gets the
# job done for now. Do something smarter if it starts mattering
ssh $TARGET sudo systemctl stop journal
rsync journal main.go index.html $TARGET:/srv/journal/
ssh $TARGET sudo systemctl start journal
