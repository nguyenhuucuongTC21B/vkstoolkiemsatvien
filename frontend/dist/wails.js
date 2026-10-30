// wails.js stub for browser preview.
// When the app runs inside Wails, this file is REPLACED by the real
// runtime that exposes window.go.<package>.<Service>.<Method> bindings
// generated from the Go Bind field declarations in main.go.
//
// In a real Wails build (`wails build`), the build step:
//   1. Generates frontend/wailsjs/go/main/App.js, go/modules/Calc.js, ...
//   2. Generates frontend/wailsjs/runtime.js
//   3. Indexes them via frontend/wailsjs/index.js
// Our app.js wcall() helper uses `window.go.<path>` directly, which
// works with both the real bindings and this stub.
(function() {
  window.wails = window.wails || {};
  // No-op for browser preview — real Wails runtime installs itself here.
  console.log("[wails.js stub] Running in browser preview mode. Real Wails bindings will be used in production.");
})();
