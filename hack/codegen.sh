#!/bin/bash -e

# Copyright 2018 infinivision, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

scriptdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

${scriptdir}/generate-groups.sh \
  deepcopy,client,lister,informer \
  github.com/infinivision/hyena-operator/pkg/client \
  github.com/infinivision/hyena-operator/pkg/apis \
  "infinivision.com:v1alpha1"
