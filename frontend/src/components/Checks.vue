<template>
  <div class="container">
    <div v-if="prerequisiteInstall" class="col-286">
      <p>A prerequisite is being installed. You might see a popup asking for permissions (could just be blinking in the taskbar)</p>
    </div>
    <div v-if="!prerequisiteInstall && checkStatus !== 'Failed'" class="col-286">
      <p >{{checkStatus}}...</p>
    </div>
    <div v-if="!prerequisiteInstall && checkStatus === 'Failed'" class="col-wide">
      <div class="failureReason" v-if="checkStatus === 'Failed'">
          Checks failed:<br/>
          {{failureReason}}
      </div>
      <p v-if="!prerequisiteInstall && checkStatus === 'Failed'">
        <a class="button" @click="check">Retry</a>
      </p>
    </div>
  </div>
  
</template>

<script>


export default {
  data() {
    return {
      prerequisiteInstall: false,
      checkStatus: "Checking mining software",
      failureReason: ""
    };
  },
  mounted() {
    this.check();
    var self = this;
    wails.events.on("checkStatus",(result) => {
		  self.checkStatus = result;
	  });
    wails.events.on("prerequisiteInstall",(result) => {
      self.prerequisiteInstall = (result === "1");
    });
  },
  methods: {
    check: function() {
  	  var self = this;
	  
      window.backend.MinerCore.PerformChecks().then(result => {
          if(result === "ok") {
      			self.startMining()
		      } else {
            self.failureReason = result;
          }
      });
    },
    startMining: function() {
      var self = this;
      window.backend.MinerCore.StartMining().then(result => {
        self.$emit('mining');
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
  font-family: 'Courier New', Courier, monospace;
  color: red;
  border: 1px solid red;
  width: 600px;
  margin: 0 auto;
}
</style>
