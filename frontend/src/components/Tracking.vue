<template>
  <div class="tracking">
    <p v-if="tracking">OCM v{{version}} - <span>You are anonymously sharing usage statistics.&nbsp;</span><a @click="disableTracking">Disable</a> - <a @click="reportIssue">Report an issue</a></p>
    <p v-if="!tracking">OCM v{{version}} - <span>You are not sharing usage statistics.&nbsp;</span><a @click="enableTracking">Enable</a><span>&nbsp;these to help us improve your experience - <a @click="reportIssue">Report an issue</a></span></p>

  </div>
</template>

<script>


export default {
  data() {
    return {
      tracking: false,
      version: "dev"
    };
  },
  mounted() {
    var self = this
     window.backend.MinerCore.TrackingEnabled().then((result) => {
       self.tracking = (result === "1")
     })
     window.backend.MinerCore.GetVersion().then((result) => {
        self.version = result 
     })
  },
  methods: {
    reportIssue: function() { 
      window.backend.MinerCore.ReportIssue()
    },
    enableTracking: function() { 
      this.tracking = true
      window.backend.MinerCore.EnableTracking()
    },
    disableTracking: function() { 
      this.tracking = false
      window.backend.MinerCore.DisableTracking()
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

div.tracking {
  position: absolute;
  bottom: 10px;
  width: 100%;
  text-align: center;
  margin: 0px;
  padding: 0px;
}
div.tracking p {
  font-size: 10px;
}
div.tracking p span {
  opacity: 0.6;
}

div.tracking p a {
  opacity: 1;
  text-decoration: underline;
  cursor: pointer;
}
</style>
