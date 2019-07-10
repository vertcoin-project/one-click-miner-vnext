<template>
  <div class="container">
    <div v-if="checkStatus !== 'Failed'" class="col-286">
      <p >{{checkStatus}}...</p>
    </div>
    <div v-if="checkStatus === 'Failed'" class="col-wide">
      <div class="failureReason" v-if="checkStatus === 'Failed'">
          Checks failed:<br/>
          {{failureReason}}
      </div>
      <p v-if="checkStatus === 'Failed'">
        <a class="button" @click="check">Retry</a>
      </p>
    </div>
  </div>
  
</template>

<script>


export default {
  data() {
    return {
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
