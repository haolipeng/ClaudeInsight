#!/usr/bin/env bash

set -ex

CMD="$1"
FILE_PREFIX="/tmp/claudeinsight"
LNAME="${FILE_PREFIX}_base.log"

function test_claudeinsight() {
    timeout 20 ${CMD} watch --debug-output http 2>&1 | tee "${LNAME}" &
    sleep 10
    curl http://www.baidu.com &>/dev/null || true
    wait

    cat "${LNAME}"
    cat "${LNAME}" | grep 'www.baidu.com'
}

function main() {
    test_claudeinsight
}

main
