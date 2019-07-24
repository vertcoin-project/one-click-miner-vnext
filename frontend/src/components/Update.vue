<template>
  <div class="container">
    <div class="col-wide">
    	<p>{{ $t('update.new_version_available') }}: {{version}}</p>
      <div class="releaseNotes">
        <pre>{{notes}}</pre>
      </div>
      <p><a class="button" @click="download">{{ $t('update.download') }}</a></p>
      <p><a class="link" @click="back">{{ $t('generic.back_to_wallet') }}</a></p>
    </div>
  </div>
</template>

<script>


export default {
  data() {
    return {
      version: "",
      notes: "",
      downloadUrl: ""
    };
  },
  mounted() {
    var self = this;
	  window.backend.Backend.VersionDetails().then((result) => {
      self.version = result[0];
      self.notes = result[1];
      self.downloadUrl = result[2];

	  });
  },
  methods: {
    back: function() {
      this.$emit('back');
    },
    download: function() { 
      window.backend.Backend.OpenDownloadUrl(this.downloadUrl);
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
    
div.releaseNotes {
  height: 130px;
  overflow-y: auto;
  font-family: 'Courier New', Courier, monospace;
  color:#eee;
  border: 1px solid #eee;
  width: 600px;
  margin: 0 auto;
}
</style>
