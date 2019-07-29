import Vue from "vue";
import App from "./App.vue";
import VueI18n from "vue-i18n";

Vue.use(VueI18n);
Vue.config.productionTip = false;
Vue.config.devtools = true;

// Import all locales
import locale_en from "./i18n/en.json";
import locale_es from "./i18n/es.json";
import locale_hr from "./i18n/hr.json";
import locale_it from "./i18n/it.json";
import locale_ja from "./i18n/ja.json";
import locale_nl from "./i18n/nl.json";
import locale_pl from "./i18n/pl.json";
import locale_pt from "./i18n/pt.json";
import locale_sl from "./i18n/sl.json";
import locale_sv from "./i18n/sv.json";

import Bridge from "./wailsbridge";

Bridge.Start(() => {
  window.backend.Backend.GetLocale().then(result => {
    
    const i18n = new VueI18n({
      locale: result, // set locale
      messages : {
        en: locale_en,
        es: locale_es,
        hr: locale_hr,
        it: locale_it,
        ja: locale_ja,
        nl: locale_nl,
        pl: locale_pl,
        pt: locale_pt, 
        sl: locale_sl,
        sv: locale_sv,
      },
    });

    new Vue({
      i18n,
      render: h => h(App)
    }).$mount("#app");
  });
});
