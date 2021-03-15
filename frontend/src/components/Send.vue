<template>
  <div class="container">
    <div v-if="sendError === '' && sent === false" class="col-286">
      <p v-if="receivedBalance === '0.00 VTC'">{{ $t('sending.send_all_to') }}:</p>
      <p
        v-if="receivedBalance !== '0.00 VTC' && receivedTxCount === 1"
      >{{ $t('sending.youre_sending_x_to', { receivedBalance }) }}:</p>
      <p
        v-if="receivedBalance !== '0.00 VTC' && receivedTxCount > 1"
      >{{ $t('sending.youre_sending_x_in_y_txs_to', { receivedBalance, receivedTxCount }) }}:</p>
      <p>
        <input
          :class="{error: (error !== ''), success: (error === '' && target !== '')}"
          @blur="recalculate()"
          type="text"
          v-model="target"
          v-bind:placeholder="$t('sending.receiver_address')"
        />
      </p>
      <p v-if="error != ''" class="error">{{error}}</p>
      <p>
        <input
          type="password"
          v-model="password"
          v-bind:placeholder="$t('sending.wallet_password')"
          @keyup.enter="send"
        />
      </p>
      <p>
        <a class="button" @click="send">{{ $t('sending.send') }}</a>
      </p>
    </div>
    <div v-if="sent === true" class="col-286">
      <svg
        style="fill: #eee"
        width="57px"
        height="57px"
        viewBox="0 0 57 57"
        version="1.1"
        xmlns="http://www.w3.org/2000/svg"
        xmlns:xlink="http://www.w3.org/1999/xlink"
      >
        <!-- Generator: Sketch 55.2 (78181) - https://sketchapp.com -->
        <title>Group</title>
        <desc>Created with Sketch.</desc>
        <g id="Page-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
          <g id="Sent" transform="translate(-371.000000, -156.000000)">
            <g id="Group" transform="translate(372.000000, 157.000000)">
              <g
                id="noun_tick_1167345"
                transform="translate(13.152174, 17.336957)"
                fill="#1EA068"
                fill-rule="nonzero"
              >
                <path
                  d="M2.45108696,11.1195652 C2.0326087,10.6413043 1.25543478,10.5815217 0.777173913,11 C0.298913043,11.4184783 0.239130435,12.1956522 0.657608696,12.673913 C0.657608696,12.673913 0.717391304,12.7336957 0.717391304,12.7336957 L7.95108696,20.5652174 C8.19021739,20.8043478 8.48913043,20.923913 8.78804348,20.923913 C9.08695652,20.923913 9.38586957,20.8043478 9.625,20.5652174 L27.8586957,2.33152174 C28.3369565,1.85326087 28.3369565,1.13586957 27.8586957,0.657608696 C27.3804348,0.179347826 26.6630435,0.179347826 26.1847826,0.657608696 L8.84782609,17.9945652 L2.45108696,11.1195652 Z"
                  id="Path"
                />
              </g>
              <circle
                id="Oval"
                stroke="#1EA068"
                stroke-width="1.19999993"
                cx="27.5"
                cy="27.5"
                r="27.5"
              />
            </g>
          </g>
        </g>
      </svg>
      <p>{{ $t('sending.coins_sent') }}</p>
      <p v-if="txids.length > 1">
        {{ $t('sending.view_trans_plural') }}:
        <br />
        <a
          class="link"
          style="display: inline"
          v-for="(txid, idx) in txids"
          v-bind:key="txid"
          @click="showTx(txid)"
        >
          <span v-if="idx > 0">&nbsp;&nbsp;</span>
          #{{idx+1}}
        </a>
      </p>
      <p v-if="txids.length == 1">
        <a
          class="link"
          style="display: inline"
          v-for="txid in txids"
          v-bind:key="txid"
          @click="showTx(txid)"
        >{{ $t('sending.view_trans_singular') }}</a>
      </p>
      <p>
        <a class="link" @click="back">{{ $t('generic.back_to_wallet') }}</a>
      </p>
    </div>
    <div v-if="sendError !== ''" class="col-286">
      <p>{{ $t('sending.failed_to_send') }}</p>
      <p>{{sendError}}</p>
      <p>
        <a class="button" @click="retry">{{ $t('generic.retry') }}</a>
      </p>
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
      receivedTxCount: 0,
      error: "",
      sent: false,
      sendError: "",
      txids: []
    };
  },
  mounted() {
    var self = this;
    window.wails.Events.On("createTransactionResult", result => {
      self.receivedBalance = result.FormattedAmount;
      self.receivedTxCount = result.NumberOfTransactions;
    });
  },
  methods: {
    send: function() {
      var self = this;
      if (this.error !== "") {
        this.sent = false;
        this.txids = [];
        this.sendError = this.error;
        return;
      }

      if (this.password === "") {
        this.sent = false;
        this.txids = [];
        this.sendError = this.$t("sending.password_required");
        return;
      }

      if (this.target === "") {
        this.sent = false;
        this.txids = [];
        this.sendError = this.$t("sending.invalid_address");
        return;
      }

      window.backend.Backend.SendSweep(this.password).then(result => {
        if (result.length === 1 && result[0].length !== 64) {
          // Error!
          self.sent = false;
          self.txids = [];
          self.sendError = self.$t("sending." + result[0]);
        } else {
          self.txids = result;
          self.sent = true;
          self.sendError = "";
        }
      });
    },
    recalculate() {
      var self = this;
      this.receivedBalance = "0.00 VTC";
      this.invalidAddress = false;
      window.backend.Backend.PrepareSweep(this.target).then(result => {
        if (result !== "") {
          self.receivedBalance = "0.00 VTC";
          self.error = self.$t("sending." + result);
        } else {
          self.error = "";
        }
      });
    },
    back() {
      this.$emit("back");
    },
    retry() {
      this.password = "";
      this.receivedBalance = "0.00 VTC";
      this.error = "";
      this.sent = false;
      this.txids = [];
      this.sendError = "";
      this.recalculate();
    },
    showTx(txid) {
      window.backend.Backend.ShowTx(txid);
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
