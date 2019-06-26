<template>
  <div class="container">
    <p>Welcome to the Vertcoin One Click Miner</p>
    <div class="col-12" v-if="walletInitialized === 0">
    	<p>Enter a password to protect your mining rewards.<br/><b><u>Don't lose it!</u></b></p>
	    <p><input type="password" v-model="password" placeholder="Password" /><br/>
	    <input type="password" v-model="confirmPassword" placeholder="Confirm Password" /></p>
    </div>
    <div class="col-12" v-if="walletInitialized === 1">
      <p>Your wallet is already initialized. Click the button below to start mining again.</p>
    </div>
      <div class="col-12" v-if="walletInitialized === -1">
      <p>Checking your wallet...</p>
    </div>
    <div class="col-12" style="position: fixed; bottom: 10px" v-if="walletInitialized === 1">
      <p>
          <a @click="start">Start Mining!</a>
      </p>
    </div>
    <div class="col-12" style="position: fixed; bottom: 10px" v-if="walletInitialized === 0" >
      <p>
          <a @click="initAndStart">Start Mining!</a>
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
</style>
