<template>
  <div id="app" unselectable="on" onselectstart="return false;">
    <TabBar
      v-on:send="switchToSend"
      v-on:wallet="switchToWallet"
      v-on:settings="switchToSettings"
      v-if="((screen === 'welcome' && manualStop) || screen !== 'welcome') && (screen !== 'checks' || tabBarVisible)"
    />
    <Welcome v-if="screen === 'welcome'" v-on:start-mining="switchToChecks" />
    <Checks v-if="screen === 'checks'" v-on:mining="switchToMining" v-on:checksFailed="showTabBar" />
    <Send v-if="screen === 'send'" v-on:back="switchToMining" v-on:cancel="switchToMining" />
    <Mining v-show="screen === 'mining'" v-on:stop-mining="stopMining" />
    <Settings v-if="screen === 'settings'" v-on:committed="restartMining" />
    <Update v-if="screen === 'update'" v-on:back="restartMiningIfNotStopped" />
    <Tracking v-on:update="switchToUpdate" />
  </div>
</template>

<script>
import Welcome from "./components/Welcome.vue";
import Mining from "./components/Mining.vue";
import Checks from "./components/Checks.vue";
import Send from "./components/Send.vue";
import Settings from "./components/Settings.vue";
import TabBar from "./components/TabBar.vue";
import Tracking from "./components/Tracking.vue";
import Update from "./components/Update.vue";

import "./assets/css/main.css";

export default {
  data() {
    return {
      screen: "welcome",
      manualStop: false,
      tabBarVisible: false
    };
  },
  mounted() {
    var self = this;
    window.wails.Events.On("minerRapidFail", () => {
      window.backend.Backend.StopMining().then(() => {
        self.switchToChecks();
      });
    });
  },
  methods: {
    stopMining: function() {
      this.manualStop = true;
      this.switchToWelcome();
    },
    // Target for the wallet tab (meta between welcome (if stopped) and mining (if mining))
    switchToWallet: function() {
      var self = this;
      if (this.tabBarVisible === true) {
        this.tabBarVisible = false;
        window.backend.Backend.StopMining().then(() => {
          self.switchToChecks();
        });
      } else {
        if (this.manualStop) {
          this.switchToWelcome();
        } else {
          this.switchToMining();
        }
      }
    },
    showTabBar: function() {
      this.tabBarVisible = true;
    },
    switchToChecks: function() {
      this.screen = "checks";
    },
    switchToSettings: function() {
      this.screen = "settings";
    },
    switchToSend: function() {
      this.screen = "send";
    },
    switchToMining: function() {
      this.manualStop = false;
      this.screen = "mining";
    },
    switchToUpdate: function() {
      this.screen = "update";
    },
    switchToWelcome: function() {
      this.screen = "welcome";
    },
    restartMiningIfNotStopped: function() {
      if(this.manualStop) {
        this.switchToWelcome();
      } else {
        this.restartMining();
      }
    },
    restartMining: function() {
      var self = this;
      window.backend.Backend.StopMining().then(() => {
        self.switchToChecks();
      });
    }
  },
  name: "app",
  components: {
    Welcome,
    Mining,
    Checks,
    Send,
    TabBar,
    Tracking,
    Settings,
    Update
  }
};
</script>
