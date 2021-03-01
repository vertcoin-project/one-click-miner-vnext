import Vue from "vue";
import App from "./App.vue";
import VueI18n from "vue-i18n";
import * as Wails from "@wailsapp/runtime"

Vue.use(VueI18n);
Vue.config.productionTip = false;
Vue.config.devtools = true;

// Import all locales
import locale_da from "./i18n/da.json";
import locale_de from "./i18n/de.json";
import locale_en from "./i18n/en.json";
import locale_es from "./i18n/es.json";
import locale_fr from "./i18n/fr.json";
import locale_hi from "./i18n/hi.json";
import locale_hr from "./i18n/hr.json";
import locale_it from "./i18n/it.json";
import locale_ja from "./i18n/ja.json";
import locale_nl from "./i18n/nl.json";
import locale_no from "./i18n/no.json";
import locale_pa from "./i18n/pa.json";
import locale_pl from "./i18n/pl.json";
import locale_pt from "./i18n/pt.json";
import locale_ro from "./i18n/ro.json";
import locale_ru from "./i18n/ru.json";
import locale_sl from "./i18n/sl.json";
import locale_sv from "./i18n/sv.json";
import locale_tr from "./i18n/tr.json";
import locale_zh from "./i18n/zh.json";

Wails.Init(() => {
    window.backend.Backend.GetLocale().then(result => {

        const i18n = new VueI18n({
            locale: result, // set locale
            fallbackLocale: 'en',
            messages: {
                da: locale_da,
                de: locale_de,
                en: locale_en,
                es: locale_es,
                fr: locale_fr,
                hi: locale_hi,
                hr: locale_hr,
                it: locale_it,
                ja: locale_ja,
                nl: locale_nl,
                no: locale_no,
                pa: locale_pa,
                pl: locale_pl,
                pt: locale_pt,
                ro: locale_ro,
                ru: locale_ru,
                sl: locale_sl,
                sv: locale_sv,
				tr: locale_tr,
                zh: locale_zh,
            },
        });
        
        new Vue({
            i18n,
            render: h => h(App)
        }).$mount("#app");
    });
});
