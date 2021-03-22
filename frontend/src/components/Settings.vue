<template>
  <div class="settings-container">
    <div class="col-settings" v-if="!showWarning">
      <div class="col-settings-sub">
        <p style="text-align: left">
          <input type="checkbox" v-model="enableIntegrated" />
          {{ $t("settings.enable_integrated") }}
          <br />
          <span class="subtext">{{ $t("settings.enable_integrated_sub") }}</span>
        </p>
        <p v-if="poolID == 5" style="text-align: left">
          {{ $t("settings.payout") }}:
          <br />
          <select style="width: 100%" name="payout" v-model="payoutID" v-on:change="clearAddress">
            <option v-for="option in payouts" v-bind:value="option.id" v-bind:key="option.id">
                {{ option.name }}
            </option>
          </select>
        </p>
        <!-- TODO: Improve address validation -->
        <p v-if="poolID == 5 && payoutID != 4">
          <input type="text" style="width:90%" class="critical-input"
            v-model="customAddress"
            v-bind:placeholder="$t('settings.customAddress')"
            v-on:change="sanitizeAddress"
          />
        </p>
        <p v-if="poolID == 6" style="text-align: left">
          <br><a class="link" @click="payoutInformation">{{ $t('settings.address_validation') }}</a>
          {{ $t("settings.mine_verthash") }}
          <br />
          {{ $t("settings.paid_doge") }}
          <br><a class="link" @click="copyAddress">{{ $t('mining.copy_address') }}</a>
          <br />
        </p>
      </div>
      <div class="col-settings-sub">
<!--        <p style="text-align: left">-->
<!--          <input type="checkbox" v-model="testnet" />-->
<!--          {{ $t("settings.testnet") }}-->
<!--          <br />-->
<!--          <span class="subtext">{{ $t("settings.testnet_sub") }}</span>-->
<!--        </p>-->
        <p style="text-align: left">
          <input type="checkbox" v-model="autoStart" />
          {{ $t("settings.auto_start") }}
          <br />
          <span class="subtext">{{ $t("settings.auto_start_sub") }}</span>
        </p>
        <p style="text-align: left" >
          <input type="checkbox" v-model="debugging" />
          {{ $t("settings.enable_debug") }}
          <br />
          <span class="subtext">{{ $t("settings.enable_debug_sub") }}</span>
        </p>
        <p style="text-align: left">
          {{ $t("settings.pool") }}:
          <br />
          <select style="width: 100%" name="pool" v-model="poolID">
            <option v-for="option in pools" v-bind:value="option.id" v-bind:key="option.id">
                {{ option.name }}
            </option>
          </select>
        </p>
      </div>
    </div>
    <div class="col-300 height-100" v-if="!showWarning">
      <p>
        <a class="button" @click="save">{{ $t("settings.save_n_restart") }}</a>
      </p>
    </div>
    <div class="col-286" v-if="showWarning">
      <div class="warning" v-if="closedSourceMiner && showWarning">
        <p>{{ $t("settings.closed_source_warning") }}</p>
      </div>
       <p>
        <a class="button" @click="toggleWarning">{{ $t("generic.close") }}</a>
      </p>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      closedSourceMiner: false,
      debugging: false,
      autoStart: false,
      enableIntegrated: true,
      showWarning: false,
      testnet: false,
      poolID: -1,
      payoutID: -1,
      pools: [],
      payouts: [],
      customAddress: "",
      address:"",

    };
  },
  created() {
    var self = this;
    window.backend.Backend.GetClosedSource().then(result => {
      self.closedSourceMiner = result;
      window.backend.Backend.GetAutoStart().then(result => {
        self.autoStart = result;
        window.backend.Backend.GetDebugging().then(result => {
          self.debugging = result;
           window.backend.Backend.GetTestnet().then(result => {
            self.testnet = result;
            window.backend.Backend.GetEnableIntegrated().then(result => {
              self.enableIntegrated = result;
              window.backend.Backend.GetPools().then(result => {
                self.pools = result;
                window.backend.Backend.GetPool().then(result => {
                  self.poolID = result;
                  window.backend.Backend.GetPayouts().then(result => {
                    self.payouts = result;
                    window.backend.Backend.GetPayout().then(result => {
                      self.payoutID = result;
                      window.backend.Backend.GetCustomAddress().then(result => {
                        self.customAddress = result;
                        window.backend.Backend.Address().then(result => {
                          self.address = result;
                        })
                      })
                    })
                  })
                })
              })
            })
          });
        });
      });
    });
    
    
   
    
  },
  methods: {
    toggleWarning: function() {
      this.showWarning = !this.showWarning;
      var self = this;
      setTimeout(() => { self.showWarning = false; }, 5000);
    },
    sanitizeAddress: function() {
      var self = this;
      self.customAddress = self.customAddress.trim();
    },
    clearAddress: function() {
      var self = this;
      self.customAddress = "";
    },
    payoutInformation: function() {
      window.backend.Backend.PayoutInformation();
    },
    copyAddress: function() {
      var textArea = document.createElement("textarea");
      textArea.value = this.address;
      // textArea.style.display = "none";
      // Avoid scrolling to bottom
      textArea.style.top = "0";
      textArea.style.left = "0";
      textArea.style.position = "fixed";
    
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
    
      try {
        document.execCommand('copy');
      } catch(e) {
        // ignore
      }
    
      document.body.removeChild(textArea);
    },
    save: function() {
      var self = this;
      window.backend.Backend.SetClosedSource(this.closedSourceMiner).then(() => {
          window.backend.Backend.SetDebugging(self.debugging).then(() => {
            window.backend.Backend.SetAutoStart(self.autoStart).then(() => {
              window.backend.Backend.SetTestnet(self.testnet).then(() => {
                window.backend.Backend.SetEnableIntegrated(self.enableIntegrated).then( () => {
                  window.backend.Backend.SetPool(self.poolID).then(() => {
                    window.backend.Backend.SetCustomAddress(self.customAddress).then(() => {
                      window.backend.Backend.SetPayout(self.payoutID).then(() => {
                        self.$emit("committed");
                      });
                    });
                  });
                });
              });
            });
          });
        }
      );
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
div.warning {
  border: 2px solid #d0a000;
  color: #d0a000;
  width: 100%;
  padding: 5px 10px;
  text-align: justify;
  line-height: 10pt;
  font-size: 10pt;
}
a.warning {
  display: block; 
  float:right;
  color: #d0a000;
  cursor: pointer;
  text-decoration: underline;
}
div.warning p {
  margin: 0px;
  padding: 0px;
}
span.subtext {
  opacity: 0.6;
  font-size: 8pt;
}
.critical-input::placeholder {
  color: red;
}
a.link {
  font-size: 14px;
  opacity: 0.6;
  text-decoration: underline;
  cursor: pointer;
  text-align: left;
}
</style>
