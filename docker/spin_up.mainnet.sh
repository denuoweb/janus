#!/bin/sh
docker-compose -f ${GOPATH}/src/github.com/htmlcoin/janus/docker/quick_start/docker-compose.mainnet.yml up -d
# sleep 3 #executing too fast causes some errors
# docker cp ${GOPATH}/src/github.com/htmlcoin/janus/docker/fill_user_account.sh htmlcoin_testchain:.
# docker exec htmlcoin_mainnet /bin/sh -c ./fill_user_account.sh
