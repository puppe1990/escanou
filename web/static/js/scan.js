(function () {
  const SCANNER_ID = "cais-scanner";
  const LOOKUP_FORM_ID = "scan-lookup-form";
  const START_BTN_ID = "cais-scan-start";
  const STATUS_ID = "cais-scan-status";
  const LAST_MARKET_KEY = "cais_last_supermarket_id";

  let scanner = null;
  let scanning = false;

  function setStatus(msg) {
    const el = document.getElementById(STATUS_ID);
    if (el) el.textContent = msg;
  }

  function setStartLabel(label) {
    const el = document.getElementById("cais-scan-start-label");
    if (el) el.textContent = label;
  }

  function restoreLastSupermarket() {
    const saved = localStorage.getItem(LAST_MARKET_KEY);
    if (!saved) return;
    document.querySelectorAll("[data-cais-supermarket-select]").forEach((sel) => {
      if (sel.querySelector(`option[value="${saved}"]`)) {
        sel.value = saved;
      }
    });
  }

  function bindSupermarketPersistence() {
    document.addEventListener("change", (e) => {
      const sel = e.target.closest("[data-cais-supermarket-select]");
      if (sel && sel.value) {
        localStorage.setItem(LAST_MARKET_KEY, sel.value);
      }
    });
  }

  function submitBarcode(code) {
    const form = document.getElementById(LOOKUP_FORM_ID);
    if (!form) return;
    const input = form.querySelector('input[name="barcode"]');
    if (input) input.value = code;
    if (window.htmx) {
      htmx.trigger(form, "submit");
    } else {
      form.requestSubmit();
    }
  }

  async function stopScanner() {
    if (!scanner || !scanning) return;
    try {
      await scanner.stop();
    } catch (_) {
      /* already stopped */
    }
    scanning = false;
    setStartLabel("Abrir câmera");
    setStatus("Câmera pausada — toque para escanear de novo");
  }

  async function requestCameraAccess() {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      throw new Error("Câmera não suportada neste navegador");
    }
    const stream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: { ideal: "environment" } },
      audio: false,
    });
    stream.getTracks().forEach((t) => t.stop());
  }

  async function startScanner() {
    const el = document.getElementById(SCANNER_ID);
    const btn = document.getElementById(START_BTN_ID);
    if (!el || !btn || typeof Html5Qrcode === "undefined") {
      setStatus("Scanner indisponível — use a busca manual ou Demo Rápido");
      return;
    }

    if (scanning) {
      await stopScanner();
      return;
    }

    btn.disabled = true;
    setStatus("Solicitando acesso à câmera…");

    try {
      await requestCameraAccess();

      if (!scanner) {
        scanner = new Html5Qrcode(SCANNER_ID);
      }

      const config = { fps: 10, qrbox: { width: 250, height: 120 }, aspectRatio: 1.5 };
      const cameras = await Html5Qrcode.getCameras();
      if (!cameras.length) {
        setStatus("Nenhuma câmera encontrada — use Demo Rápido à direita");
        return;
      }

      const back = cameras.find((c) => /back|rear|traseira|environment/i.test(c.label));
      const cameraId = (back || cameras[cameras.length - 1]).id;

      let scanned = false;
      await scanner.start(
        cameraId,
        config,
        (decoded) => {
          if (scanned) return;
          scanned = true;
          setStatus("Código lido — buscando produto…");
          scanner
            .stop()
            .catch(() => {})
            .finally(() => {
              scanning = false;
              setStartLabel("Abrir câmera");
              submitBarcode(decoded);
            });
        },
        () => {}
      );

      scanning = true;
      setStartLabel("Parar câmera");
      setStatus("Aponte ao código de barras dentro da moldura");
    } catch (err) {
      const name = err && err.name ? err.name : "";
      if (name === "NotAllowedError" || name === "PermissionDeniedError") {
        setStatus("Câmera bloqueada — clique no ícone de cadeado na barra de endereço e permita");
      } else if (name === "NotFoundError") {
        setStatus("Nenhuma câmera encontrada — use Demo Rápido à direita");
      } else {
        setStatus((err && err.message) || "Erro ao abrir câmera — use Demo Rápido");
      }
      console.warn("cais scan:", err);
    } finally {
      btn.disabled = false;
    }
  }

  function init() {
    restoreLastSupermarket();
    bindSupermarketPersistence();
    document.body.addEventListener("htmx:afterSwap", (e) => {
      if (e.detail.target && e.detail.target.id === "scan-result") {
        restoreLastSupermarket();
      }
    });

    const btn = document.getElementById(START_BTN_ID);
    if (btn) {
      btn.addEventListener("click", () => startScanner());
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();