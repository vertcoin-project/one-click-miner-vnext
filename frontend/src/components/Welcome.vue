<template>
  <div class="container">
    <div class="col-286" v-if="walletInitialized === 0">
    	<p>Make a password. Don't lose it.</p>
	    <p><input type="password" v-model="password" placeholder="Password" /></p>
	    <p><input type="password" v-model="confirmPassword" placeholder="Confirm Password" /></p>
      <p><a class="button" @click="initAndStart">Start Mining!</a></p>
    </div>
    <div class="col-286" v-if="walletInitialized === 1">
      <p>Click the button below to start mining again.</p>
      <p>
          <a class="button" @click="start">Start Mining!</a>
      </p>
    </div>
  </div>
</template>

<script>


export default {
  data() {
    return {
      password: "",
	  confirmPassword: "",
      walletInitialized: -1
    };
  },
  created() {
	  var self = this;
	  window.backend.MinerCore.WalletInitialized().then(result => {
      self.walletInitialized = result;
      if(self.walletInitialized === 1 && !(self.$parent.manualStop === true)) {
        self.start();
      }
	  });
  },
  methods: {
    initAndStart: function() {
      if(this.password === '') {
        alert("Password cannot be empty");
        return;
      }
      if(this.password !== this.confirmPassword) {
        alert("Passwords do not match");
        return
      }
      
      var self = this;
	  
      window.backend.MinerCore.InitWallet(this.password).then(result => {
        if(result !== true){
			    alert('Something went initializing the wallet.');
        } else {
          self.start()
        }
      });
    },
    start: function() {
      this.$emit('start-mining');
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
