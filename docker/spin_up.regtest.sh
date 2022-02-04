#!/bin/sh
docker-compose -f ${GOPATH}/src/github.com/htmlcoin/janus/docker/quick_start/docker-compose.regtest.yml up -d
sleep 3 #executing too fast causes some errors
docker cp ${GOPATH}/src/github.com/htmlcoin/janus/docker/fill_user_account.sh htmlcoin_regtest:.
docker exec htmlcoin_regtest /bin/sh -c ./fill_user_account.sh
