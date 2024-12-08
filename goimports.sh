#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


# Fix order of imports in go files.
RED='\033[0;31m'
NC='\033[0m' # No Color

if [ $(uname -s) == "Darwin" ]; then

find internal -name \*.go | xargs -P 16 -I{} sed -i '' -e '
/^import/,/)/ {
/^$/ d
}
' {}

elif [ $(uname -s) == "Linux" ]; then

find internal -name \*.go | xargs -P 8 -I{} sed -i -e '
/^import/,/)/ {
/^$/ d
}
' {} 

else
    echo -e  "${RED}Please run the command in linux dev container or mac.${NC}"
    exit 1
fi


find internal -name \*.go | xargs -I{} goimports -local terraform-provider-oodle -w {}

