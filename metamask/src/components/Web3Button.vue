<template>
  <div class="hello">
    <div v-if="web3Detected">
      <b-button v-if="htmlcoinConnected">Connected to HTMLCOIN</b-button>
      <b-button v-else-if="connected" v-on:click="connectToHtmlcoin()">Connect to HTMLCOIN</b-button>
      <b-button v-else v-on:click="connectToWeb3()">Connect</b-button>
    </div>
    <b-button v-else>No Web3 detected - Install metamask</b-button>
  </div>
</template>

<script>
let HTMLCOINMainnet = {
  chainId: '0x115C', // 4444
  chainName: 'HTMLCOIN Mainnet',
  rpcUrls: ['https://janus.htmlcoin.com/api/'],
  blockExplorerUrls: ['https://info.htmlcoin.com/'],
  iconUrls: [
    'https://htmlcoin.com/images/metamask_icon.svg',
    'https://htmlcoin.com/images/metamask_icon.png',
  ],
  nativeCurrency: {
    decimals: 18,
    symbol: 'HTMLCOIN',
  },
};
let HTMLCOINTestNet = {
  chainId: '0x115D', // 4445
  chainName: 'HTMLCOIN Testnet',
  rpcUrls: ['https://testnet-janus.htmlcoin.com/api/'],
  blockExplorerUrls: ['https://testnet-explorer.htmlcoin.com/'],
  iconUrls: [
    'https://htmlcoin.com/images/metamask_icon.svg',
    'https://htmlcoin.com/images/metamask_icon.png',
  ],
  nativeCurrency: {
    decimals: 18,
    symbol: 'HTMLCOIN',
  },
};
let HTMLCOINRegTest = {
  chainId: '0x115E', // 4446
  chainName: 'HTMLCOIN Regtest',
  rpcUrls: ['https://localhost:24889'],
  // blockExplorerUrls: ['https://testnet-explorer.htmlcoin.com/'],
  iconUrls: [
    'https://htmlcoin.com/images/metamask_icon.svg',
    'https://htmlcoin.com/images/metamask_icon.png',
  ],
  nativeCurrency: {
    decimals: 18,
    symbol: 'HTML',
  },
};
let config = {
  "0x115C": HTMLCOINMainnet,
  "0x115D": HTMLCOINTestNet,
  "0x115E": HTMLCOINRegTest,
};

export default {
  name: 'Web3Button',
  props: {
    msg: String,
    connected: Boolean,
    htmlcoinConnected: Boolean,
  },
  computed: {
    web3Detected: function() {
      return !!this.Web3;
    },
  },
  methods: {
    getChainId: function() {
      return window.htmlcoin.chainId;
    },
    isOnHtmlcoinChainId: function() {
      let chainId = this.getChainId();
      return chainId == HTMLCOINMainnet.chainId || chainId == HTMLCOINTestNet.chainId;
    },
    connectToWeb3: function(){
      if (this.connected) {
        return;
      }
      let self = this;
      window.htmlcoin.request({ method: 'eth_requestAccounts' })
        .then(() => {
          console.log("Emitting web3Connected event");
          let htmlcoinConnected = self.isOnHtmlcoinChainId();
          let currentlyHtmlcoinConnected = self.htmlcoinConnected;
          self.$emit("web3Connected", true);
          if (currentlyHtmlcoinConnected != htmlcoinConnected) {
            console.log("ChainID matches HTMLCOIN, not prompting to add network to web3, already connected.");
            self.$emit("htmlcoinConnected", true);
          }
        })
        .catch((e) => {
          console.log("Connecting to web3 failed", arguments, e);
        })
    },
    connectToHtmlcoin: function() {
      console.log("Connecting to Htmlcoin, current chainID is", this.getChainId());

      let self = this;
      let htmlcoinConfig = config[this.getChainId()] || HTMLCOINTestNet;
      console.log("Adding network to Metamask", htmlcoinConfig);
      window.htmlcoin.request({
        method: "wallet_addEthereumChain",
        params: [htmlcoinConfig],
      })
        .then(() => {
          self.$emit("htmlcoinConnected", true);
        })
        .catch(() => {
          console.log("Adding network failed", arguments);
        })
    },
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
