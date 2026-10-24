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
    b.GetClosedSource(),
    b.GetAutoStart(),
    b.GetDebugging(),
    b.GetTestnet(),
    b.GetEnableIntegrated(),
    b.GetPools(),
    b.GetPool(),
    b.GetLanguages(),
    b.GetLanguage(),
    b.GetSystemLocale()
  ]).then(function(r) {
    last = {
      closedSourceMiner: r[0],
      autoStart: r[1],
      debugging: r[2],
      testnet: r[3],
      enableIntegrated: r[4],
      pools: r[5],
      poolID: r[6],
      languages: r[7],
      language: r[8],
      systemLocale: r[9]
    };
    return last;
  });
}
