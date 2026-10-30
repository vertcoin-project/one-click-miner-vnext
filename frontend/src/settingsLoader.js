// Reads everything the settings screen shows from the backend. The calls don't depend
// on each other, so they run in parallel, and the last result is kept so the screen can
// show it right away the next time it is opened.
var last = null;

export function cachedSettings() {
  return last;
}

export function forgetSettings() {
  last = null;
}

export function loadSettings() {
  var b = window.backend.Backend;
  return Promise.all([
    b.GetAutoStart(),
    b.GetDebugging(),
    b.GetEnableIntegrated(),
    b.GetPools(),
    b.GetPool(),
    b.GetLanguages(),
    b.GetLanguage(),
    b.GetSystemLocale()
  ]).then(function(r) {
    last = {
      autoStart: r[0],
      debugging: r[1],
      enableIntegrated: r[2],
      pools: r[3],
      poolID: r[4],
      languages: r[5],
      language: r[6],
      systemLocale: r[7]
    };
    return last;
  });
}
