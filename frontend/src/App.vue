<template>
  <div id="app" unselectable="on" onselectstart="return false;" >
    <TabBar v-on:send="switchToSend" v-on:wallet="switchToMining" :data-screen="screen" v-if="screen !== 'welcome' && screen !== 'checks'" />
    <Welcome v-if="screen === 'welcome'" v-on:start-mining="switchToChecks"/>
    <Checks v-if="screen === 'checks'" v-on:mining="switchToMining"/>
    <Send v-if="screen === 'send'" v-on:back="switchToMining" v-on:cancel="switchToMining"/>
    <Mining v-if="screen === 'mining'" v-on:stop-mining="stopMining"  />
  </div>
</template>

<script>
import Welcome from "./components/Welcome.vue";
import Mining from "./components/Mining.vue";
import Checks from "./components/Checks.vue";
import Send from "./components/Send.vue";
import TabBar from "./components/TabBar.vue";
import "./assets/css/main.css";

export default {
  data() {
    return {
      screen : "welcome",
      manualStop: false
    };
  },
  methods: {
    stopMining: function() {
        this.manualStop = true;
        this.switchToWelcome();
    },
    switchToChecks: function() {
		this.screen = 'checks';
    },
    switchToSend: function() {
        this.screen = 'send';
	},
	switchToMining: function() {
        this.screen = 'mining';
	},
	switchToWelcome: function() {
		this.screen = 'welcome';
	},
  },
  name: "app",
  components: {
    Welcome,
    Mining,
    Checks,
    Send,
    TabBar
  }
};
</script>
