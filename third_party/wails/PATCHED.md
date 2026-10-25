# Wails v1.16.9, patched

This is github.com/wailsapp/wails v1.16.9 (MIT, see LICENSE) with two changes, used
through the `replace` directive in the top-level go.mod. Only the packages the app
compiles are here: the root package, `cmd`, `lib/...` and `runtime` with the assets
it embeds. Tests, the CLI (`cmd/wails`), the project templates and the website are
left out.

## Unique sync callback IDs

`lib/renderer/webview.go`, `evalJSSync`: the callback ID was

    fmt.Sprintf("syncjs:%d:%d", time.Now().Unix(), rand.Intn(9999))

At start-up Wails calls `evalJSSync` once per bound Go method, all within the same
second. With ~50 bound methods two IDs collide about one launch in seven (the
birthday problem over 9999 values). The handler registered for the first ID stays
registered, so the later call's callback runs both handlers, `wg.Done()` is called
twice and the app dies with `panic: sync: negative WaitGroup counter`. Go 1.20+
seeds the global rand source randomly, which is why it is intermittent.

The random number is now a counter from `atomic.AddUint64`, so every ID is unique.

## Show the window once the page is ready

On Windows, `webview_init` in `lib/renderer/webview/webview.h` showed the window
straight after writing the start page into the browser control, before the control
had painted. The control starts out white, so on some launches the window faded in
white before the app's dark background appeared.

The window is now created hidden on Windows. `webview_show` (exposed to Go as
`Show()` in `lib/renderer/webview/webview.go`) shows it, and `lib/renderer/webview.go`
calls it once Wails emits `wails:ready`, after the bindings, CSS and app JavaScript
are in, or after 3 seconds if that never happens. On Linux and macOS the window is
shown in `webview_init` as before and `webview_show` does nothing.

## Checking the copy

`third_party/verify-wails.sh` downloads v1.16.9 into an empty module cache, which
Go checks against the public checksum database (sum.golang.org), then confirms
that every file here matches it byte for byte, apart from this file and the files
`third_party/wails.patch` changes, which must be v1.16.9's with the patch applied.
CI runs it on every change to `third_party` or `go.mod`.

## Updating Wails

Copy the same files from the new version's module and apply the changes again,
then update the version and hash in `third_party/verify-wails.sh`. Or drop this
folder, the patch, the script and the `replace` line if upstream has fixed it.
