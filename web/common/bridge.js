// Базовый мост площадки window.tnk9xPlatform (контракт — web/README.md).
// Это поведение обычной веб-страницы: игра приостанавливается на скрытой
// вкладке, звук молчит, пока игра приостановлена, сохранения лежат
// в localStorage, язык — язык браузера, рекламы и покупок нет. Мост портала (web/<портал>/platform.js)
// подключается после этого файла и переопределяет нужные части.
// Методы моста не бросают исключений: syscall/js превращает их в panic
(function () {
  "use strict";

  // Причины приостановки: игра стоит, пока есть хоть одна
  const reasons = new Set();
  const suspendListeners = [];

  // Go-аудио (oto) создаёт свой AudioContext при запуске wasm: обёртка
  // конструктора запоминает контексты, чтобы глушить их целиком —
  // на скрытой вкладке кадры игры не идут, а звук играет дальше
  const audioContexts = [];
  const NativeAudioContext = window.AudioContext || window.webkitAudioContext;
  const nativeResume = NativeAudioContext && NativeAudioContext.prototype.resume;

  if (NativeAudioContext) {
    class TrackedAudioContext extends NativeAudioContext {
      constructor(...args) {
        super(...args);
        audioContexts.push(this);
      }

      // oto возобновляет звук по жестам пользователя; пока игра
      // приостановлена, такие запросы игнорируются
      resume() {
        if (reasons.size > 0) {
          return Promise.resolve();
        }
        return nativeResume.call(this);
      }
    }
    window.AudioContext = TrackedAudioContext;
    if (window.webkitAudioContext) {
      window.webkitAudioContext = TrackedAudioContext;
    }
  }

  function resumeAudio() {
    for (const context of audioContexts) {
      if (context.state === "suspended") {
        nativeResume.call(context).catch(() => {});
      }
    }
  }

  function syncAudio() {
    if (reasons.size > 0) {
      for (const context of audioContexts) {
        context.suspend().catch(() => {});
      }
      return;
    }
    resumeAudio();
  }

  // Браузеры (iOS) возобновляют звук только по жесту: после
  // приостановки он вернётся с первым касанием или нажатием
  for (const type of ["touchend", "keyup", "mouseup"]) {
    document.addEventListener(type, () => {
      if (reasons.size === 0) {
        resumeAudio();
      }
    });
  }

  function notifySuspend() {
    for (const listener of suspendListeners) {
      try {
        listener(platform.suspended);
      } catch (_) {
        // слушатель моста портала не должен ломать базовый мост
      }
    }
  }

  // Канвас игры получает фокус клавиатуры: внутри iframe портала
  // ebiten не фокусирует его сам
  function focusCanvas() {
    const canvas = document.querySelector("canvas");
    if (canvas && document.activeElement !== canvas) {
      canvas.focus();
    }
  }
  window.addEventListener("focus", focusCanvas);
  new MutationObserver((_, observer) => {
    if (document.querySelector("canvas")) {
      observer.disconnect();
      focusCanvas();
    }
  }).observe(document.documentElement, { childList: true, subtree: true });

  const platform = {
    // Состояние, которое игра читает каждый кадр
    suspended: false,
    suspensions: 0,
    rewardAvailable: false,
    purchasesAvailable: false,
    // Язык интерфейса площадки; игра читает его один раз при запуске
    language: "",

    // Подготовка площадки до запуска wasm
    init() {
      return Promise.resolve();
    },
    ready() {},
    gameplay() {},
    intermission() {},
    requestReward() {},
    rewardStatus() {
      return "";
    },
    catalog() {
      return [];
    },
    purchases() {
      return [];
    },
    requestPurchase() {},
    purchaseStatus() {
      return "";
    },
    consumePurchase() {},

    storage: {
      getItem(key) {
        try {
          return window.localStorage.getItem(key);
        } catch (_) {
          return null;
        }
      },
      setItem(key, value) {
        try {
          window.localStorage.setItem(key, value);
          return true;
        } catch (_) {
          return false;
        }
      },
    },

    // Для мостов порталов: причины приостановки и подписка на неё
    suspend(reason) {
      if (reasons.has(reason)) {
        return;
      }
      reasons.add(reason);
      if (reasons.size === 1) {
        platform.suspended = true;
        platform.suspensions++;
        syncAudio();
        notifySuspend();
      }
    },
    resume(reason) {
      if (reasons.delete(reason) && reasons.size === 0) {
        platform.suspended = false;
        syncAudio();
        notifySuspend();
      }
    },
    addSuspendListener(listener) {
      suspendListeners.push(listener);
    },
    // Для мостов порталов: язык площадки, он же язык страницы
    setLanguage(language) {
      if (typeof language !== "string" || language === "") {
        return;
      }
      platform.language = language;
      document.documentElement.lang = language;
    },
  };

  platform.setLanguage(navigator.language);

  function syncVisibility() {
    if (document.hidden) {
      platform.suspend("hidden");
    } else {
      platform.resume("hidden");
    }
  }
  document.addEventListener("visibilitychange", syncVisibility);
  syncVisibility();

  window.tnk9xPlatform = platform;
})();
