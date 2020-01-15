<template>
  <div class="container">
    <div class="col-286" v-if="alreadyRunning === false && walletInitialized === 0">
      <p>{{ $t("welcome.makeapassword") }}</p>
      <p class="error" v-if="error !== ''">{{error}}</p>
      <p><input type="password" v-model="password" v-bind:placeholder="$t('welcome.password')" /></p>
      <p><input type="password" v-model="confirmPassword" v-bind:placeholder="$t('welcome.confirmpassword')" @keyup.enter="initAndStart" /></p>
      <p><a class="button" @click="initAndStart">{{ $t("welcome.startmining") }}</a></p>
    </div>
    <div class="col-286" v-if="alreadyRunning === false && walletInitialized === 1">
      <p>{{ $t("welcome.click_button_to_start") }}</p>
      <p>
          <a class="button" @click="start">{{ $t("welcome.startmining") }}</a>
      </p>
    </div>
    <div class="col-286" v-if="alreadyRunning === true">
      <p>{{ $t("welcome.alreadyrunning") }}</p>
      <p>
          <a class="button" @click="close">{{ $t("generic.close") }}</a>
      </p>
    </div>
  </div>
</template>

<script>



export default {
  data() {
    return {
      error: "",
      alreadyRunning: false,
      password: "",
      confirmPassword: "",
      walletInitialized: -1
    };
  },
  created() {
    var self = this;
    window.backend.Backend.AlreadyRunning().then(result => {
      self.alreadyRunning = result;
      if (!result) {
        window.backend.Backend.WalletInitialized().then(result => {
          self.walletInitialized = result;
          if(self.walletInitialized === 1 && !(self.$parent.manualStop === true)) {
            self.start();
          }
        });

      }
    })
  },
  methods: {
    close: function() {
      window.backend.Backend.Close();
    },
    initAndStart: function() {
      this.error = '';
      if(this.password === '') {
        this.error = this.$t("welcome.password_cannot_be_empty");
        return;
      }
      if(this.password !== this.confirmPassword) {
        this.error = this.$t("welcome.password_mismatch");
        return
      }
      
      var self = this;

      window.backend.Backend.InitWallet(this.password).then(result => {
        if(result !== true){
          this.error = this.$t("welcome.error_initializing");
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
  p.error {
    color: red;
  }
</style>
