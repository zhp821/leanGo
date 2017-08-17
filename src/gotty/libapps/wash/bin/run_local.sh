#!/bin/bash
# Copyright (c) 2012 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

cd "$(readlink -f "$(dirname "$0")/..")"

if [ -z $DISPLAY ]; then
  export DISPLAY=":0.0"
fi

if [ -z "$CHROME_TEST_PROFILE" ]; then
  CHROME_TEST_PROFILE=$HOME/.config/google-chrome-run_local
fi

mkdir -p $CHROME_TEST_PROFILE

./bin/mkdeps.sh

google-chrome --load-and-launch-app="$(pwd)" \
  --allow-file-access-from-files \
  --user-data-dir=$CHROME_TEST_PROFILE \
  &>/dev/null </dev/null &
