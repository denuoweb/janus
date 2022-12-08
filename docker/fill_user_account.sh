#!/bin/sh
repeat_until_success () {
    echo Running command - "$@"
    i=0
    until $@
    do
        echo Command failed with exit code - $?
        if [ $i -gt 10 ]; then
            echo Giving up running command - "$@"
            return
        fi
        echo Sleeping $i seconds
        sleep $i
        echo Retrying
        i=`expr $i + 1`
    done
    echo Command finished successfully
}

# Htmlcoin v22.1 no longer auto creates a default wallet, we need to create one to import keys and such
RANDOM=$(date +%s)
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd createwallet "createdWallet"$(( $RANDOM % 129 ))

#import private keys and then prefund them
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd importprivkey "cMbgxCJrTYUqgcmiC1berh5DFrtY1KeU4PXZ6NZxgenniF1mXCRk" address1 # addr=hUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW hdkeypath=m/88'/0'/1'
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd importprivkey "cRcG1jizfBzHxfwu68aMjhy78CpnzD9gJYZ5ggDbzfYD3EQfGUDZ" address2 # addr=hLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf hdkeypath=m/88'/0'/2'
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd importprivkey "cV79qBoCSA2NDrJz8S3T7J8f3zgkGfg4ua4hRRXfhbnq5VhXkukT" address3 # addr=hTCCy8qy7pW94EApdoBjYc1vQ2w68UnXPi
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd importprivkey "cV93kaaV8hvNqZ711s2z9jVWLYEtwwsVpyFeEZCP6otiZgrCTiEW" address4 # addr=hWMi6ne9mDQFatRGejxdDYVUV9rQVkAFGp
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd importprivkey "cVPHpTvmv3UjQsZfsMRrW5RrGCyTSAZ3MWs1f8R1VeKJSYxy5uac" address5 # addr=hLcshhsRS6HKeTKRYFdpXnGVZxw96QQcfm
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd importprivkey "cTs5NqY4Ko9o6FESHGBDEG77qqz9me7cyYCoinHcWEiqMZgLC6XY" address6 # addr=hW28njWueNpBXYWj2KDmtFG2gbLeALeHfV

echo Finished importing accounts
echo Seeding accounts
echo Seeding hcavTSEVe31NLdXyfq925GzGp8yN5QnS6a
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 2 hcavTSEVe31NLdXyfq925GzGp8yN5QnS6a
htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 2 ha1W7VnwtJPoFDoNjxxGdDHBtsRKDpjW8c
htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 1000 haofg5zZVyvPmWgGL6YdVAyRTKWd3MjZ4A
# address1
echo Seeding hUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 1000 hUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW
# address2
echo Seeding hLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 1000 hLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf
# address3
echo Seeding hTCCy8qy7pW94EApdoBjYc1vQ2w68UnXPi
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 500 hTCCy8qy7pW94EApdoBjYc1vQ2w68UnXPi
# address4
echo Seeding hWMi6ne9mDQFatRGejxdDYVUV9rQVkAFGp
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 250 hWMi6ne9mDQFatRGejxdDYVUV9rQVkAFGp
# address5
echo Seeding hLcshhsRS6HKeTKRYFdpXnGVZxw96QQcfm
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 100 hLcshhsRS6HKeTKRYFdpXnGVZxw96QQcfm
# address6
echo Seeding hW28njWueNpBXYWj2KDmtFG2gbLeALeHfV
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 100 hW28njWueNpBXYWj2KDmtFG2gbLeALeHfV
# playground pet shop dapp
echo Seeding 0xCca81b02942D8079A871e02BA03A3A4a8D7740d2
repeat_until_success htmlcoin-cli -rpcuser=htmlcoin -rpcpassword=testpasswd generatetoaddress 2 hcDWPLgdY9pTv3cKLkaMPvqjukURH3Qudy
echo Finished importing and seeding accounts
