<template>
  <div class="container">
    <div class="col-12">
	    	<p class="header">Spendable Balance:</p>
	    	<p class="spendableBalance">{{balance}} VTC 
					&nbsp;
					<a class="tiny" @click="refreshBalance">
						<svg width="16" height="16" version="1.1" id="Capa_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
								viewBox="0 0 458.186 458.186" style="enable-background:new 0 0 458.186 458.186;" xml:space="preserve">
							<g>
								<g>
									<path style="fill: #ffffff" d="M445.651,201.95c-1.485-9.308-10.235-15.649-19.543-14.164c-9.308,1.485-15.649,10.235-14.164,19.543
										c0.016,0.102,0.033,0.203,0.051,0.304c17.38,102.311-51.47,199.339-153.781,216.719c-102.311,17.38-199.339-51.47-216.719-153.781
										S92.966,71.232,195.276,53.852c62.919-10.688,126.962,11.29,170.059,58.361l-75.605,25.19
										c-8.944,2.976-13.781,12.638-10.806,21.582c0.001,0.002,0.002,0.005,0.003,0.007c2.976,8.944,12.638,13.781,21.582,10.806
										c0.003-0.001,0.005-0.002,0.007-0.002l102.4-34.133c6.972-2.322,11.675-8.847,11.674-16.196v-102.4
										C414.59,7.641,406.949,0,397.523,0s-17.067,7.641-17.067,17.067v62.344C292.564-4.185,153.545-0.702,69.949,87.19
										s-80.114,226.911,7.779,310.508s226.911,80.114,310.508-7.779C435.905,339.799,457.179,270.152,445.651,201.95z"/>
								</g>
							</g>
						</svg>
					</a>
					&nbsp;
					<a class="tiny" @click="sendMoney">
						<svg width="16" height="16" viewBox="0 0 512 512.00004" xmlns="http://www.w3.org/2000/svg"><path style="fill: #ffffff" d="m511.824219 255.863281-233.335938-255.863281v153.265625h-27.105469c-67.144531 0-130.273437 26.148437-177.753906 73.628906-47.480468 47.480469-73.628906 110.609375-73.628906 177.757813v107.347656l44.78125-49.066406c59.902344-65.628906 144.933594-103.59375 233.707031-104.457032v153.253907zm-481.820313 179.003907v-30.214844c0-59.132813 23.027344-114.730469 64.839844-156.542969s97.40625-64.839844 156.539062-64.839844h57.105469v-105.84375l162.734375 178.4375-162.734375 178.441407v-105.84375h-26.917969c-94.703124 0-185.773437 38.652343-251.566406 106.40625zm0 0"/></svg>
					</a>
				</p>
				<p class="immatureBalance" v-if="balanceImmature != '0.00000000'"><small>(<b>{{balanceImmature}} VTC</b> still maturing)</small></p>
				<p class="poolBalance" v-if="balancePendingPool != '0.00000000'"><small>(<b>{{balancePendingPool}} VTC</b> pending pool payout)</small></p>
				<p class="spacer">&nbsp;</p>
				<p class="header">Expected Earnings (24h):</p>
	    	<p class="earning">~{{avgearn}} ({{hashrate}})</p>
    </div>
	<!--<div class="col-6">
	    	<p><b><u>Your Graphics Card(s):</u></b></p>
	    	<p>{{gpu}}</p>
		<p><b><u>Your Wallet Address:</u></b></p>
	    	<p>{{wallet}}</p>
    </div>-->
		<div class="col-12" style="position: fixed; bottom: 10px">
			<p>
					<a @click="stop">Stop Mining</a>
			</p>
		</div>
  </div>
	
</template>

<script>
export default {
  data() {
    return {
      hashrate: "0 MH/s",
			avgearn:"0.00 VTC",
	  	gpu: "Unknown",
	  	wallet: "Unknown",
	  	balance: "0.00000000",
			balanceImmature: "0.00000000",
			balancePendingPool: "0.00000000"
    };
  },
  mounted() {
	var self = this;
	wails.events.on("hashRate",(result) => {
		self.hashrate = result;
	});
	wails.events.on("avgEarnings",(result) => {
		self.avgearn = result;
	});
	wails.events.on("balance",(result) => {
		self.balance = result;
	});
	wails.events.on("balanceImmature",(result) => {
		self.balanceImmature = result;
	});
	wails.events.on("balancePendingPool",(result) => {
		self.balancePendingPool = result;
	});
  window.backend.MinerCore.GetGPUs().then(result => {
		self.gpu = result[0];
	});
	 window.backend.MinerCore.Address().then(result => {
		self.wallet = result;
	});
  },
  methods: {
    stop: function() {
			var self = this;
			window.backend.MinerCore.StopMining().then(result => {
				self.$emit('stop-mining');
			});
		},
		refreshBalance: function() { 
			window.backend.MinerCore.RefreshBalance();
		},
		sendMoney: function() { 
			this.$emit('send')
		}
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
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
a.tiny:hover {
  opacity: 1.0;
  transition: 500ms;
}
a.tiny {
		opacity: 0.6;
    background: #048652;
    display: inline;
    color: white;
		z-index: 500;
		padding: 5px;
		margin: 0px;
    cursor: pointer;
    border-radius: 5px;
}

p.spendableBalance {
	margin: 0;
	padding: 0;
	font-size: 24px;
}
p.immatureBalance, p.poolBalance {
	margin: 0;
	padding: 0;
	font-size: 10pt;
}
p.spacer {
	padding: 0px;
	margin: 5px;
}
p.header {
	margin-bottom: 0;
	padding-bottom: 5px;
	font-weight: bold;
	text-decoration: underline;
}
p.earning {
	margin: 0;
	padding: 0;
	font-size: 20px;
}

</style>
