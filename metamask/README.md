# Simple VUE project to switch to HTMLCOIN network via Metamask

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

### wallet_addEthereumChain
```
// request account access
window.htmlcoin.request({ method: 'eth_requestAccounts' })
    .then(() => {
        // add chain
        window.htmlcoin.request({
            method: "wallet_addEthereumChain",
            params: [{
                {
                    chainId: '0x115D',
                    chainName: 'Htmlcoin Testnet',
                    rpcUrls: ['https://localhost:24889'],
                    blockExplorerUrls: ['https://testnet-explorer.htmlcoin.com/'],
                    iconUrls: [
                        'https://htmlcoin.com/images/metamask_icon.svg',
                        'https://htmlcoin.com/images/metamask_icon.png',
                    ],
                    nativeCurrency: {
                        decimals: 18,
                        symbol: 'HTML',
                    },
                }
            }],
        }
    });
```

# Known issues
- Metamask requires https for `rpcUrls` so that must be enabled
  - Either directly through Janus with `--https-key ./path --https-cert ./path2` see [SSL](../README.md#ssl)
  - Through the Makefile `make docker-configure-https && make run-janus-https`
  - Or do it yourself with a proxy (eg, nginx)
