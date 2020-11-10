<template>
  <div class="tabBar">
    <div class="tabs">
      <div
        :class="{tab : true, active: $parent.screen === 'mining' || $parent.screen === 'welcome'}"
      >
        <a @click="wallet">{{ $t('tabbar.wallet') }}</a>
      </div>
      <div :class="{tab : true, active: $parent.screen === 'send'}">
        <a @click="send">{{ $t('tabbar.send_coins') }}</a>
      </div>
      <div :class="{tab : true, active: $parent.screen === 'settings'}">
        <a @click="settings">{{ $t('tabbar.settings') }}</a>
      </div>
    </div>
    <div style="float: right" v-if="testnet">
      <div class="testnet">TESTNET</div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      testnet: false
    };
  },
  mounted() {
    var self = this;
    window.backend.Backend.GetTestnet().then(result => {
      self.testnet = result;
    })
  },
  methods: {
    send: function() {
      this.$emit("send");
    },
    wallet: function() {
      this.$emit("wallet");
    },
    settings: function() {
      this.$emit("settings");
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
div.tabBar {
  position: relative;
  height: 58px;
  background-color: rgba(255, 255, 255, 0.1);
  width: 100%;
  margin: 0px;
  padding: 0px;
}

div.tabs {
  width: 500px;
  position: absolute;
  left: 50%;
  top: 50%;
  height: 20px;
  margin-left: -250px;
  margin-top: -10px;
  padding: 0px;
}

div.tab {
  width: 33%;
  margin: 0px;
  padding: 0px;
  float: left;
  color: white;
  opacity: 0.6;
  text-align: center;
}

div.tab.active {
  opacity: 1;
}

a {
  cursor: pointer;
}


div.testnet {
  height: 38px;
  color: #ff0000;
  border: 1px solid #ff0000;
  font-weight: bold;
  padding-left: 10px;
  padding-right: 10px;
  font-size: 20px;
  margin-top: 10px;
  margin-right: 20px;
  line-height: 38px;
  text-align: center;
  width: 100px;
}
</style>
