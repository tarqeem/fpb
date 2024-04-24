package main

import "github.com/chromedp/cdproto/page"

// see: https://intoli.com/blog/not-possible-to-block-chrome-headless/
const script = `(function(w, n, wn) {
  // Pass the Webdriver Test.
 Object.defineProperty(navigator, 'fonts', { value: ['Helvetica', 'Helvetica'] });
 Intl.DateTimeFormat().resolvedOptions = () => ({ timeZone: 'America/New_York', formatMatcher: 'best fit' });



  // Pass the Plugins Length Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'plugins', {
    // This just needs to have length > 0 for the current test,
    // but we could mock the plugins too if necessary.
    get: () => [1, 2, 3, 4, 5],
  });

  // Pass the Languages Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'languages', {
    get: () => ['en-US', 'en'],
  });

  // Pass the Chrome Test.
  // We can mock this in as much depth as we need for the test.
  w.chrome = {
    runtime: {},
  };

  // Pass the Permissions Test.
  const originalQuery = wn.permissions.query;
  return wn.permissions.query = (parameters) => (
    parameters.name === 'notifications' ?
      Promise.resolve({ state: Notification.permission }) :
      originalQuery(parameters)
  );

})(window, navigator, window.navigator);`

var buf []byte
var scriptID page.ScriptIdentifier
