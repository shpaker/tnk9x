// Загрузчик игры: wasm скачивается параллельно с подготовкой площадки
// (не дольше initTimeoutMs — игра запускается и без неё), затем
// стартует Go. Перед этим файлом подключаются мост площадки
// и wasm_exec.js
(function () {
  "use strict";

  const initTimeoutMs = 5000;
  const status = document.getElementById("status");
  const go = new Go();

  function initPlatform() {
    const init = Promise.resolve()
      .then(() => window.tnk9xPlatform.init())
      .catch(() => {});
    const timeout = new Promise((resolve) => setTimeout(resolve, initTimeoutMs));
    return Promise.race([init, timeout]);
  }

  async function instantiate() {
    if (WebAssembly.instantiateStreaming) {
      try {
        return await WebAssembly.instantiateStreaming(
          fetch("tnk9x.wasm"), go.importObject);
      } catch (_) { /* неверный MIME у локального сервера — fallback ниже */ }
    }
    const buf = await (await fetch("tnk9x.wasm")).arrayBuffer();
    return await WebAssembly.instantiate(buf, go.importObject);
  }

  Promise.all([instantiate(), initPlatform()])
    .then(([result]) => {
      if (status) {
        status.remove();
      }
      go.run(result.instance);
    })
    .catch((err) => {
      if (status) {
        status.textContent = "failed to load: " + err;
      }
    });
})();
