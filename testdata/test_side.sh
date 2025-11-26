#!/usr/bin/env bash
. $(dirname "$0")/common.sh
set -ex

CMD="$1"
DOCKER_REGISTRY="$2"
FILE_PREFIX="/tmp/claudeinsight"
CLIENT_MATCH_LNAME="${FILE_PREFIX}_side_client_match.log"
CLIENT_NONMATCH_LNAME="${FILE_PREFIX}_side_client_nonmatch.log"

function test_side_client_match() {
    if [ -z "$DOCKER_REGISTRY" ]; then
        IMAGE_NAME="nginx:1.23.1"
    else
        IMAGE_NAME=$DOCKER_REGISTRY"/library/nginx:1.23.1"
    fi
    docker pull "$IMAGE_NAME"

    cname='test-nginx-side'
    docker rm -f $cname || true
    cid1=$(docker run --name $cname -p 8080:80 -d "$IMAGE_NAME")
    export cid1
    echo $cid1

    timeout 30 ${CMD} watch --debug-output http --side client 2>&1 | tee "${CLIENT_MATCH_LNAME}" &
    sleep 10
    for i in $(seq 1 5); do
        curl -s http://127.0.0.1:8080/ > /dev/null || true
        sleep 0.3
    done
    wait

    cat "${CLIENT_MATCH_LNAME}"
    cat "${CLIENT_MATCH_LNAME}" | grep '\[side\]=client'
    check_patterns_not_in_file "${CLIENT_MATCH_LNAME}"  '\[side\]=server'

    docker rm -f $cid1 || true
}

function test_side_client_nonmatch() {
    if [ -z "$DOCKER_REGISTRY" ]; then
        IMAGE_NAME="nginx:1.23.1"
    else
        IMAGE_NAME=$DOCKER_REGISTRY"/library/nginx:1.23.1"
    fi
    docker pull "$IMAGE_NAME"

    cname='test-nginx-side'
    docker rm -f $cname || true
    cid1=$(docker run --name $cname -p 8080:80 -d "$IMAGE_NAME")
    export cid1
    echo $cid1

    timeout 30 ${CMD} watch --debug-output http --side server 2>&1 | tee "${CLIENT_NONMATCH_LNAME}" &
    sleep 10
    for i in $(seq 1 20); do
        curl -s http://127.0.0.1:8080/ > /dev/null || true
        sleep 0.3
    done
    wait

    cat "${CLIENT_NONMATCH_LNAME}"
    cat "${CLIENT_NONMATCH_LNAME}" | grep '\[side\]=server'
    check_patterns_not_in_file "${CLIENT_NONMATCH_LNAME}"  '\[side\]=client'

    docker rm -f $cid1 || true
}

function main() {
    test_side_client_match
    test_side_client_nonmatch
}

main
