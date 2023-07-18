#!/bin/sh
docker-compose -f ${GOPATH}/src/github.com/denuoweb/janus/docker/quick_start/docker-compose.mainnet.yml up -d
# sleep 3 #executing too fast causes some errors
# docker cp ${GOPATH}/src/github.com/denuoweb/janus/docker/fill_user_account.sh htmlcoin_testchain:.
# docker exec htmlcoin_mainnet /bin/sh -c ./fill_user_account.sh
