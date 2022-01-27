#!/usr/bin/env bats
# Copyright (c) 2021 SIGHUP s.r.l All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.


load "./../../helper"

OS="linux"
if [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
fi

@test "Prepare temporal ssh key" {
    info
    ssh_keys(){
        cp ./automated-tests/e2e-tests/vsphere/sshkey.pub /tmp/sshkey.pub
    }
    run ssh_keys
    if [[ $status -ne 0 ]]; then
        echo "$output" >&3
    fi
    [ "$status" -eq 0 ]
}

@test "Cluster destroy" {
    info
    destroy(){
        ./dist/furyctl-${OS}_${OS}_amd64/furyctl -d --debug cluster destroy --force --config ./automated-tests/e2e-tests/vsphere/cluster.yml -w ./automated-tests/e2e-tests/vsphere/cluster
    }
    run destroy

    if [[ $status -ne 0 ]]; then
        echo "$output" >&3
        echo "  TERRAFORM LOGS:" >&3
        cat ./automated-tests/e2e-tests/vsphere/cluster/logs/terraform.logs >&3
    fi
    [ "$status" -eq 0 ]
}