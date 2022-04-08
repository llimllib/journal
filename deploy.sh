#!/usr/bin/env bash
set -euxo pipefail

# set the linux build to be transferred
mv journal journal-local
cp journal-linux journal
trap "mv journal-local journal" EXIT

# it's dumb to have the whole server stopped while rsync'ing, but it gets the
# job done for now. Do something smarter if it starts mattering
#
# KEY is the location of the keyfile to use to log in
# TARGET is user@hostname
ssh -i "$KEY" "$TARGET" sudo systemctl stop llimllib-journal
rsync -e "ssh -i $KEY" journal index.html "$TARGET:/srv/journal/"
ssh -i "$KEY" "$TARGET" sudo systemctl start llimllib-journal
