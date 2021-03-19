<template>
  <div class="tracking">
    <p v-if="tracking">
      OCM v{{version}}
      <span v-if="updateAvailable">
        -
        <a @click="update">{{ $t('tracking.update_available') }}</a>
      </span> -
      <span>{{ $t('tracking.tracking_enabled') }}.&nbsp;</span>
      <a @click="disableTracking">{{ $t('tracking.disable_tracking') }}</a> -
      <a @click="reportIssue">{{ $t('tracking.report_issue') }}</a>
    </p>
    <p v-if="!tracking">
      OCM v{{version}}
      <span v-if="updateAvailable">
        -
        <a @click="update">{{ $t('tracking.update_available') }}</a>
      </span>
      <!-- </span> - -->
      <!-- <span>{{ $t('tracking.tracking_disabled') }}.&nbsp;</span>-->
      <!-- <a @click="enableTracking">{{ $t('tracking.enable_tracking') }}</a> -->
      <span>
        &nbsp;-
        <a @click="reportIssue">{{ $t('tracking.report_issue') }}</a>
      </span>
    </p>
  </div>
</template>

<script>
export default {
  data() {
    return {
      tracking: false,
      version: "dev",
      updateAvailable: false
    };
  },
  mounted() {
    var self = this;
    window.backend.Backend.TrackingEnabled().then(result => {
      self.tracking = result === "1";
    });
    window.backend.Backend.GetVersion().then(result => {
      self.version = result;
    });
    window.backend.Backend.UpdateAvailable().then(result => {
      self.updateAvailable = result;
    });
    window.wails.Events.On("updateAvailable", result => {
      self.updateAvailable = result;
    });
  },
  methods: {
    update: function() {
      this.$emit("update");
    },
    reportIssue: function() {
      window.backend.Backend.ReportIssue();
    },
    enableTracking: function() {
      this.tracking = true;
      window.backend.Backend.EnableTracking();
    },
    disableTracking: function() {
      this.tracking = false;
      window.backend.Backend.DisableTracking();
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
