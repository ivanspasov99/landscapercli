#!/bin/bash
#
# Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -e

CURRENT_DIR=$(dirname $0)
PROJECT_ROOT="${CURRENT_DIR}"/..

go install \
  -ldflags "-X github.com/gardener/landscapercli/pkg/version.gitVersion=$EFFECTIVE_VERSION \
            -X github.com/gardener/landscapercli/pkg/version.gitTreeState=$([ -z git status --porcelain 2>/dev/null ] && echo clean || echo dirty) \
            -X github.com/gardener/landscapercli/pkg/version.gitCommit=$(git rev-parse --verify HEAD)" \
  ${PROJECT_ROOT}/landscaper-cli
