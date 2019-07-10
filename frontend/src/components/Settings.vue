<template>
  <div class="container">
    <div class="col-286">
      <p style="text-align: left">
        <input type="checkbox" v-model="closedSourceMiner">Use closed-source miners<br/>
        <span class="subtext">Better hashrate, but unaudited miners that incurr a developer's fee</span>
      </p>
      <div class="warning" v-if="closedSourceMiner">
        <p>
            You have selected to use closed source miner(s). Vertcoin does 
            not endorse or support these miners. They cannot be audited on 
            their contents and could contain functions that harm your computer. 
            You are also paying a developer's fee when using these miners.
        </p>
      </div>
      <p>
          <a class="button" @click="save">Save &amp; Restart</a>
      </p>
    </div>
  </div>
</template>

<script>


export default {
  data() {
    return {
      closedSourceMiner: false,
    };
  },
  created() {
	  var self = this;
	  window.backend.MinerCore.GetClosedSource().then(result => {
      self.closedSourceMiner = result
	  });
  },
  methods: {
    save: function() {
      var self = this;
      window.backend.MinerCore.SetClosedSource(this.closedSourceMiner).then(result => {
        self.$emit('committed');
	    });
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
  div.warning p {
    margin: 0px; 
    padding: 0px;
  }
  span.subtext {
    opacity: 0.6;
    font-size: 10pt;
  }

</style>
