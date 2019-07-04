<template>
  <div class="container">
    <div v-if="sendError === '' && !sent" class="col-286">
    	<p v-if="receivedBalance === '0.00 VTC'">Send all your mined coins to:</p>
	    <p v-if="receivedBalance !== '0.00 VTC'">You're sending {{receivedBalance}} to:</p>
	    <p><input :class="{error: (error !== ''), success: (error === '' && target !== '')}" @blur="recalculate()" type="text" v-model="target" placeholder="Receiver Address" /></p>
      <p v-if="error != ''" class="error">{{error}}</p>
	    <p><input type="password" v-model="password" placeholder="Wallet Password" /></p>
      <p>
          <a class="button" @click="send">Send</a>
      </p>
    </div>
    <div v-if="sent" class="col-286">
      <svg style="fill: #048652" width="57px" height="57px" viewBox="0 0 57 57" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
          <!-- Generator: Sketch 55.2 (78181) - https://sketchapp.com -->
          <title>Group</title>
          <desc>Created with Sketch.</desc>
          <g id="Page-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
              <g id="Sent" transform="translate(-371.000000, -156.000000)">
                  <g id="Group" transform="translate(372.000000, 157.000000)">
                      <g id="noun_tick_1167345" transform="translate(13.152174, 17.336957)" fill="#1EA068" fill-rule="nonzero">
                          <path d="M2.45108696,11.1195652 C2.0326087,10.6413043 1.25543478,10.5815217 0.777173913,11 C0.298913043,11.4184783 0.239130435,12.1956522 0.657608696,12.673913 C0.657608696,12.673913 0.717391304,12.7336957 0.717391304,12.7336957 L7.95108696,20.5652174 C8.19021739,20.8043478 8.48913043,20.923913 8.78804348,20.923913 C9.08695652,20.923913 9.38586957,20.8043478 9.625,20.5652174 L27.8586957,2.33152174 C28.3369565,1.85326087 28.3369565,1.13586957 27.8586957,0.657608696 C27.3804348,0.179347826 26.6630435,0.179347826 26.1847826,0.657608696 L8.84782609,17.9945652 L2.45108696,11.1195652 Z" id="Path"></path>
                      </g>
                      <circle id="Oval" stroke="#1EA068" stroke-width="1.19999993" cx="27.5" cy="27.5" r="27.5"></circle>
                  </g>
              </g>
          </g>
      </svg>
      <p>Your coins are sent</p>
      <p><a class="link" @click="showTx">Show transaction</a></p>
      <p><a class="link" @click="back">Back to wallet</a></p>
    </div>
    <div v-if="sendError !== ''" class="col-286">
      <p>Failed to send your coins</p>
      <p>{{sendError}}</p>
      <p><a class="button" @click="retry">Retry</a></p>
    </div>
  </div>
</template>

<script>


export default {
  data() {
    return {
      invalidAddress: false,
      target: "",
      password: "",
      receivedBalance: "0.00 VTC",
      error : "",
      sent : false,
      sendError: "",
      txid : "",
    };
  },
  mounted() {
    var self = this;
    wails.events.on("createTransactionResult",(result) => {
			self.receivedBalance = result;
		});
  },
  methods: {
    send: function() {
      var self = this;
      if(this.error !== '') {
        this.sent = false;
        this.txid = '';
        this.sendError = this.error;
        return
      }

      if(this.password === "") {
        this.sent = false;
        this.txid = '';
        this.sendError = "Wallet password is required";
        return
      }

      if(this.target === "") {
        this.sent = false;
        this.txid = '';
        this.sendError = "Invalid address";
        return
      }

      window.backend.MinerCore.SendSweep(this.password).then(result => {
        if(result.length == 64) { // TXID!
          self.txid = result;
          self.sent = true;
          self.sendError = '';
        } else {
          self.sent = false;
          self.txid = '';
          self.sendError = result;
        }
      });
    },
    recalculate() {
      var self = this;
      this.receivedBalance = "0.00 VTC";
      this.invalidAddress = false;
      window.backend.MinerCore.PrepareSweep(this.target).then(result => {
        if(result !== "") {
          self.receivedBalance = "0.00 VTC";
          self.error = result;
        } else {
          self.error = "";
        }
      });
    },
    back() {
      this.$emit('back')
    },
    retry() {
      this.target = "";
      this.password = "";
      this.receivedBalance = "0.00 VTC";
      this.error =  "";
      this.sent = false;
      this.txid = "";
      this.sendError = "";
    },
    showTx() {
      window.backend.MinerCore.ShowTx(this.txid);
    }

  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>


</style>
