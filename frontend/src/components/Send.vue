<template>
  <div class="container">
    <h3><u>Send coins</u></h3>
    <div class="col-12">
    	<p>Use this page to send your coins to another wallet.</p>
	    <input type="text" v-model="target" placeholder="Receiver Address" />
    </div>
     <div class="col-6" style="position: fixed; left: 10px; bottom: 10px">
      <p>
          <a @click="cancel">Cancel</a>
      </p>
    </div>
    <div class="col-6" style="position: fixed; right: 10px; bottom: 10px">
      <p>
          <a @click="send">Send</a>
      </p>
    </div>
  </div>
</template>

<script>


export default {
  data() {
    return {
      target: "",
    };
  },
  mounted() {
    
  },
  methods: {
    send: function() {
  	  this.$emit('sent');
    },
    cancel: function() {
      this.$emit('cancel');
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
input {
  font-size: 1.3em;
  border-color: white;
  background-color: #121212;
  color: white;
  border: 1px solid white;
  border-radius: 5px;
  padding: 15px;
  width: 600px;
  margin: 0 auto;
}
a:hover {
  opacity: 1.0;
  transition: 500ms;
}
a {
	opacity: 0.6;
    font-size: 18px;
    background: #048652;
    max-width: 300px;
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
