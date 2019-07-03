<template>
  <div class="tracking">
    <p v-if="tracking"><span>You are anonymously sharing usage statistics.&nbsp;</span><a @click="disableTracking">Disable</a></p>
    <p v-if="!tracking"><span>You are not sharing usage statistics.&nbsp;</span><a @click="enableTracking">Enable</a><span>&nbsp;these to help us improve your experience</span></p>
  </div>
</template>

<script>


export default {
  data() {
    return {
      tracking: false
    };
  },
  mounted() {
    var self = this
     window.backend.MinerCore.TrackingEnabled().then((result) => {
       self.tracking = (result === "1")
     })
  },
  methods: {
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
