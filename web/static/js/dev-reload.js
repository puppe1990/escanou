(function () {
  var lastStamp = null;
  var pollMs = 1000;

  function poll() {
    fetch("/dev/reload", { cache: "no-store" })
      .then(function (res) {
        if (!res.ok) {
          throw new Error("reload check failed");
        }
        return res.text();
      })
      .then(function (stamp) {
        if (lastStamp && stamp !== lastStamp) {
          window.location.reload();
          return;
        }
        lastStamp = stamp;
      })
      .catch(function () {
        // air restart or server down — pick up new stamp when it returns
        lastStamp = null;
      });
  }

  poll();
  setInterval(poll, pollMs);
})();
