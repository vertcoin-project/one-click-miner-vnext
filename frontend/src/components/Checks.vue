<template>
  <div class="container">
    <div v-if="prerequisiteInstall" class="col-286">
      <p>{{ $t("checks.prerequisite") }}</p>
    </div>
    <div v-if="!prerequisiteInstall && checkStatus !== 'Failed'" class="col-286">
      <p>{{checkStatus === null ? $t("checks.checking_mining_software") : (checkStatus === 'Failed' ? 'Failed' : $t("checks." + checkStatus)) }}</p>
      <div class="verthashProgress" v-if="verthashProgress !== 0">
        <div class="progressBar">
          <div class="progress" v-bind:style="{width: verthashProgress + '%'}">&nbsp;</div>
        </div>
        <div class="progressText">{{Math.floor(verthashProgress)}}%</div>
      </div>
    </div>
    <div v-if="!prerequisiteInstall && checkStatus === 'Failed'" class="col-wide">
      <div class="failureReason" v-if="checkStatus === 'Failed'">
        {{ $t("checks.checks_failed") }}:
        <br />&nbsp;<br />
        {{failureReason}}
      </div>
      <p v-if="!prerequisiteInstall && checkStatus === 'Failed'">
        <a class="button" @click="check">{{ $t('generic.retry') }}</a>
      </p>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      prerequisiteInstall: false,
      checkStatus: null,
      failureReason: "",
      verthashProgress: 0,
    };
  },
  mounted() {
    this.check();
    var self = this;
    window.wails.Events.On("checkStatus", result => {
      self.checkStatus = result;
    });
    window.wails.Events.On("prerequisiteInstall", result => {
      self.prerequisiteInstall = result === "1";
    });
    window.wails.Events.On("verthashProgress", result => {
      self.verthashProgress = result;
    })
  },
  methods: {
    check: function() {
      var self = this;

      window.backend.Backend.PerformChecks().then(result => {
        if (result === "ok") {
          self.startMining();
        } else {
          self.$emit("checksFailed");
          self.failureReason = result;
        }
      });
    },
    startMining: function() {
      var self = this;
      window.backend.Backend.StartMining().then(() => {
        self.$emit("mining");
      });
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
div.failureReason {
  height: 200px;
  overflow-y: auto;
  font-family: "Courier New", Courier, monospace;
  color: white;
  border: 1px solid white;
  width: 600px;
  margin: 0 auto;
}

div.verthashProgress {
  margin: 0 auto;
  width: 200px;
}

div.progressBar {
  border: 1px solid #048652;
  height: 10px;
  width: 100%;
  margin: 0px;
  padding: 0px;
  margin-bottom: 10px;
}

div.progress {
  float:left;
  background-color: #048652;
  margin: 0px;
  padding: 0px;
  height: 10px; 
}
</style>

