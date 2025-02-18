module.exports = {
    networks: {
      development: {
        host: "127.0.0.1",
        port: 24888, //Switch to 24888 for local HTTP Server, look at Makefile run-janus
        network_id: "*",
        gasPrice: "0x64"
      },
      ganache: {
        host: "127.0.0.1",
        port: 8545,
        network_id: "*"
      },
      testnet: {
        host: "testnet.htmlcoin.com",
        port: 24889,
        network_id: "*",
        from: "0x7926223070547d2d15b2ef5e7383e541c338ffe9",
        gasPrice: "0x64"
      }
    },
    compilers: {
      solc: {
        version: "^0.6.12",
      }
    },
  }
