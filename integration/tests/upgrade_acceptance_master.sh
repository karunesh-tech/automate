#!/bin/bash

#shellcheck disable=SC2034
test_name="upgrade_acceptance_master"
test_upgrades=true
test_upgrade_inspec_profiles=()
test_upgrade_begin_manifest="acceptance.json"

do_upgrade() {
    local previous_umask
    previous_umask=$(umask)
    umask 022
    do_upgrade_default
    sudo chef-automate post-major-upgrade migrate --data=PG -y
    umask "$previous_umask"
}