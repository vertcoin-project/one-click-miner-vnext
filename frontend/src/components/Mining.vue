<template>
  <div class="container">
    <div class="col-286">
      <p class="header">{{$t('mining.spendable_balance')}}:</p>
      <p class="spendableBalance">
        <a class="tiny" @click="refreshBalance">
          <svg
            width="16"
            height="16"
            version="1.1"
            id="Capa_1"
            xmlns="http://www.w3.org/2000/svg"
            xmlns:xlink="http://www.w3.org/1999/xlink"
            x="0px"
            y="0px"
            viewBox="0 0 458.186 458.186"
            style="enable-background:new 0 0 458.186 458.186;"
            xml:space="preserve"
          >
            <g>
              <g>
                <path
                  style="fill: #048652"
                  d="M445.651,201.95c-1.485-9.308-10.235-15.649-19.543-14.164c-9.308,1.485-15.649,10.235-14.164,19.543
										c0.016,0.102,0.033,0.203,0.051,0.304c17.38,102.311-51.47,199.339-153.781,216.719c-102.311,17.38-199.339-51.47-216.719-153.781
										S92.966,71.232,195.276,53.852c62.919-10.688,126.962,11.29,170.059,58.361l-75.605,25.19
										c-8.944,2.976-13.781,12.638-10.806,21.582c0.001,0.002,0.002,0.005,0.003,0.007c2.976,8.944,12.638,13.781,21.582,10.806
										c0.003-0.001,0.005-0.002,0.007-0.002l102.4-34.133c6.972-2.322,11.675-8.847,11.674-16.196v-102.4
										C414.59,7.641,406.949,0,397.523,0s-17.067,7.641-17.067,17.067v62.344C292.564-4.185,153.545-0.702,69.949,87.19
										s-80.114,226.911,7.779,310.508s226.911,80.114,310.508-7.779C435.905,339.799,457.179,270.152,445.651,201.95z"
                />
              </g>
            </g>
          </svg>
        </a>
        &nbsp;{{balance}} VTC
      </p>
      <p class="immatureBalance" v-if="balanceImmature != '0.00000000'">
        (
        <span style="opacity: 1">{{balanceImmature}} VTC</span>
        {{$t('mining.still_maturing')}})
      </p>
      <p class="poolBalance" v-if="balancePendingPool != '0.00000000'">
        (
        <span style="opacity: 1">{{balancePendingPool}} VTC</span>
        {{$t('mining.pending_pool_payout')}})
      </p>
      <p class="spacer">&nbsp;</p>
      <p v-if="runningMiners === 0">{{$t('mining.waiting_for_miners')}}</p>
      <p v-if="runningMiners > 0" class="header">{{$t('mining.expected_earnings_24h')}}:</p>
      <p
        v-if="runningMiners > 0 && hashrate !== '0.00 MH/s'"
        class="earning"
      >~{{avgearn}} ({{hashrate}})</p>
      <p
        v-if="runningMiners > 0 && hashrate === '0.00 MH/s'"
        class="earning"
      >{{$t('mining.estimating')}}{{spinner}}</p>
      <p>
        <a class="button" v-if="stopping">{{spinner}}</a>
        <a class="button" @click="stop" v-if="!stopping">{{$t('mining.stop_mining')}}</a>
      </p>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      hashrate: "0.00 MH/s",
      avgearn: "0.00 VTC",
      netHash: "",
      gpu: "",
      wallet: "",
      balance: "0.00000000",
      balanceImmature: "0.00000000",
      balancePendingPool: "0.00000000",
      runningMiners: 0,
      spinner: "...",
      stopping: false,
    };
  },
  mounted() {
    var self = this;
    window.setInterval(() => {
      var newSpinner = self.spinner + ".";
      if (newSpinner.length > 5) {
        newSpinner = ".";
      }
      self.spinner = newSpinner;
    }, 1000);
    window.wails.Events.On("hashRate", result => {
      self.hashrate = result;
    });
    window.wails.Events.On("runningMiners", result => {
      self.runningMiners = result;
    });
    window.wails.Events.On("networkHashRate", result => {
      self.netHash = result;
    });
    window.wails.Events.On("avgEarnings", result => {
      self.avgearn = result;
    });
    window.wails.Events.On("balance", result => {
      self.balance = result;
    });
    window.wails.Events.On("balanceImmature", result => {
      self.balanceImmature = result;
    });
    window.wails.Events.On("balancePendingPool", result => {
      self.balancePendingPool = result;
    });
    window.backend.Backend.RefreshBalance();
    window.backend.Backend.RefreshHashrate();
    window.backend.Backend.RefreshRunningState();
  },
  methods: {
    stop: function() {
      var self = this;
      this.stopping = true;
      window.backend.Backend.StopMining().then(() => {
        self.$emit("stop-mining");
      });
    },
    refreshBalance: function() {
      window.backend.Backend.RefreshBalance();
    },
    sendMoney: function() {
      this.$emit("send");
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
a.tiny:hover {
  opacity: 1;
  transition: 500ms;
}
a.tiny {
  opacity: 0.6;
  background: transparent;
  display: inline;
  z-index: 500;
  margin: 0px;
  cursor: pointer;
  border: 0px;
}

p.spendableBalance,
p.earning {
  margin: 0;
  padding: 0;
  font-size: 20px;
}
p.immatureBalance,
p.poolBalance,
p.netHash {
  margin: 0;
  padding: 0;
  font-size: 12px;
  opacity: 0.6;
}
p.spacer {
  padding: 0px;
  margin: 5px;
}
p.header {
  margin-bottom: 0;
  padding-bottom: 5px;
  opacity: 0.6;
}
</style>
