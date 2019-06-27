<template>
  <div class="container">
    <p v-if="checkStatus !== 'Failed'">{{checkStatus}}...</p>
    <div class="failureReason" v-if="checkStatus === 'Failed'">
      Checks failed:<br/>
      {{failureReason}}
    </div>
    <div class="col-12" style="position: fixed; bottom: 10px" v-if="checkStatus === 'Failed'" >
      <p>
          <a @click="check">Retry</a>
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
input {
  font-size: 1em;
  border-color: white;
  background-color: #121212;
  color: white;
  border: 1px solid white;
  border-radius: 5px;
  padding: 4px;
}
a:hover {
  opacity: 1.0;
  transition: 500ms;
}
a {
	opacity: 0.6;
    font-size: 18px;
    background: #048652;
    max-width: 420px;
    margin: 0 auto;
    display: block;
    color: white;
    z-index: 500;
    margin-top: 1.5em;
    padding: 22px 21px;
    box-shadow: 0px 3px 4px rgba(0, 0, 0, 0.15);
    cursor: pointer;
    font-weight: 400 !important;
    text-align: center;
    border-radius: 5px;
}
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
