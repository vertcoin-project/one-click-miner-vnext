<template>
  <div class="settings-container">
    <div class="col-settings" v-if="!showWarning">
      <div class="col-settings-sub">
        <p style="text-align: left" >
          <input type="checkbox" v-model="debugging" />
          {{ $t("settings.enable_debug") }}
          <br />
          <span class="subtext">{{ $t("settings.enable_debug_sub") }}</span>
        </p>
        <p style="text-align: left">
          <input type="checkbox" v-model="autoStart" />
          {{ $t("settings.auto_start") }}
          <br />
          <span class="subtext">{{ $t("settings.auto_start_sub") }}</span>
        </p>
      </div>
      <div class="col-settings-sub">
        <!-- <p style="text-align: left">
          <input type="checkbox" v-model="testnet" />
          {{ $t("settings.testnet") }}
          <br />
          <span class="subtext">{{ $t("settings.testnet_sub") }}</span>
        </p> -->
        <p style="text-align: left">
          <input type="checkbox" v-model="enableIntegrated" />
          {{ $t("settings.enable_integrated") }}
          <br />
          <span class="subtext">{{ $t("settings.enable_integrated_sub") }}</span>
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
    <div class="col-286 height-100" v-if="!showWarning">
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
import { GetAutoStart, GetClosedSource, GetDebugging, GetEnableIntegrated, GetPool, GetPools, GetTestnet, SetAutoStart, SetClosedSource, SetDebugging, SetEnableIntegrated, SetPool, SetTestnet } from '../wailsjs/go/backend/Backend'

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
      pools: [],
    };
  },
  created() {
    var self = this;
    GetClosedSource().then(result => {
      self.closedSourceMiner = result;
      GetAutoStart().then(result => {
        self.autoStart = result;
        GetDebugging().then(result => {
          self.debugging = result;
          GetTestnet().then(result => {
          self.testnet = result;
            GetPools().then(result => {
              self.pools = result;
              GetPool().then(result => {
                self.poolID = result;
                GetEnableIntegrated().then(result => {
                  self.enableIntegrated = result;
                });
              });
            });
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
    save: function() {
      var self = this;
      SetClosedSource(this.closedSourceMiner).then(() => {
          SetDebugging(self.debugging).then(() => {
            SetAutoStart(self.autoStart).then(() => {
              SetTestnet(self.testnet).then(() => {
                SetPool(self.poolID).then(() => {
                  SetEnableIntegrated(self.enableIntegrated).then( () => {
                    self.$emit("committed");
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
</style>
