<template>
  <div class="container">
    <div class="col-286">
      <p style="text-align: left">
        <input type="checkbox" v-model="debugging">{{ $t("settings.enable_debug") }}<br/>
        <span class="subtext">{{ $t("settings.enable_debug_sub") }}</span>
      </p>
       <p style="text-align: left">
        <input type="checkbox" v-model="autoStart">{{ $t("settings.auto_start") }}<br/>
        <span class="subtext">{{ $t("settings.auto_start_sub") }}</span>
      </p>
      <p style="text-align: left">
        <input type="checkbox" v-model="closedSourceMiner">{{ $t("settings.closed_source") }}<br/>
        <span class="subtext">{{ $t("settings.closed_source_sub") }}</span>
      </p>
      <div class="warning" v-if="closedSourceMiner">
        <p>
            {{ $t("settings.closed_source_warning") }}
        </p>
      </div>
      <p>
          <a class="button" @click="save">{{ $t("settings.save_n_restart") }}</a>
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
    };
  },
  created() {
	  var self = this;
	  window.backend.Backend.GetClosedSource().then(result => {
      self.closedSourceMiner = result
    });
     window.backend.Backend.GetAutoStart().then(result => {
      self.autoStart = result
    });
    window.backend.Backend.GetDebugging().then(result => {
      self.debugging = result
    });
  },
  methods: {
    save: function() {
      var self = this;
      window.backend.Backend.SetClosedSource(this.closedSourceMiner).then(result => {
        window.backend.Backend.SetDebugging(self.debugging).then(result => {
          window.backend.Backend.SetAutoStart(self.autoStart).then(result => {
            self.$emit('committed');
	        });
	      });
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
    font-size: 8pt;
  }

</style>
