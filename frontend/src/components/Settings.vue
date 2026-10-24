<template>
  <div class="settings-container">
    <div class="col-settings" v-if="!showWarning">
      <div class="col-settings-sub" v-bind:style="{ visibility: loaded ? 'visible' : 'hidden' }">
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
        <p style="text-align: left">
          {{ $t("settings.language") }}:
          <br />
          <select style="width: 100%" name="language" v-model="language" @change="changeLanguage">
            <!-- In the system's language, so the way back is readable from any other -->
            <option value="">{{ $t("settings.language_system", systemLocale) }}</option>
            <option v-for="option in languages" v-bind:value="option.code" v-bind:key="option.code">
                {{ option.name }}
            </option>
          </select>
        </p>
      </div>
      <div class="col-settings-sub" v-bind:style="{ visibility: loaded ? 'visible' : 'hidden' }">
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
      <p class="bottom-button">
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
import { loadSettings, cachedSettings, forgetSettings } from "../settingsLoader.js";

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
      language: "",
      languages: [],
      systemLocale: "en",
      // False until the first values arrive, so a half loaded screen (empty pool list)
      // is never shown
      loaded: false
    };
  },
  created() {
    var self = this;
    var apply = function(settings) {
      for (var key in settings) {
        self[key] = settings[key];
      }
      self.loaded = true;
    };

    // Show the previous values right away and refresh them in the background
    var cached = cachedSettings();
    if (cached) {
      apply(cached);
    }
    loadSettings()
      .then(apply)
      .catch(function() {
        self.loaded = true;
      });
  },
  methods: {
    toggleWarning: function() {
      this.showWarning = !this.showWarning;
      var self = this;
      setTimeout(() => { self.showWarning = false; }, 5000);
    },
    // The language applies and is saved as soon as it is picked, without Save & Restart
    changeLanguage: function() {
      var self = this;
      window.backend.Backend.SetLanguage(this.language).then(() => {
        window.backend.Backend.GetLocale().then(locale => {
          self.$i18n.locale = locale;
          forgetSettings();
        });
      });
    },
    save: function() {
      var self = this;
      window.backend.Backend.SetClosedSource(this.closedSourceMiner).then(() => {
          window.backend.Backend.SetDebugging(self.debugging).then(() => {
            window.backend.Backend.SetAutoStart(self.autoStart).then(() => {
              window.backend.Backend.SetTestnet(self.testnet).then(() => {
                window.backend.Backend.SetPool(self.poolID).then(() => {
                  window.backend.Backend.SetEnableIntegrated(self.enableIntegrated).then( () => {
                    forgetSettings();
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
