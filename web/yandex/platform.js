// Мост Яндекс Игр поверх базового (web/common/bridge.js): SDK, язык
// интерфейса Яндекс Игр, разметка загрузки и геймплея, межуровневая
// реклама и реклама за награду, пауза площадки и облачные сохранения.
// Вне Яндекса (SDK не загрузился) остаётся поведение базового моста.
// Методы не бросают исключений
(function () {
  "use strict";

  const platform = window.tnk9xPlatform;
  const baseStorage = platform.storage;

  // Задержка отправки облачных сохранений: серию записей шлём одной
  const cloudSaveDelayMs = 1000;

  // Контекстное меню, выделение и перетаскивание на всей странице
  for (const type of ["contextmenu", "selectstart", "dragstart"]) {
    document.addEventListener(type, (event) => event.preventDefault());
  }

  let ysdk = null;

  // Готовность и разметка геймплея: что просила игра и что уже
  // сообщено SDK
  let readyRequested = false;
  let readySent = false;
  let gameplayRequested = false;
  let gameplayMarked = false;

  // Облачные сохранения: ключ — строка игры; null — облака нет
  let player = null;
  let cloud = null;
  let storageTouched = false;
  let cloudSaveTimer = 0;

  // Исход рекламы за награду в контракте моста
  let rewardStatus = "";

  function call(fn) {
    try {
      const result = fn();
      if (result && typeof result.catch === "function") {
        result.catch(() => {});
      }
    } catch (_) {
      // ошибка SDK не должна доходить до игры
    }
  }

  function sendReady() {
    if (!ysdk || !readyRequested || readySent) {
      return;
    }
    readySent = true;
    call(() => ysdk.features.LoadingAPI.ready());
  }

  // GameplayAPI размечает только идущую игру: приостановка площадкой
  // останавливает разметку, даже если кадры игры не идут
  function syncGameplay() {
    const active = gameplayRequested && !platform.suspended;
    if (!ysdk || active === gameplayMarked) {
      return;
    }
    gameplayMarked = active;
    call(() => active
      ? ysdk.features.GameplayAPI.start()
      : ysdk.features.GameplayAPI.stop());
  }
  platform.addSuspendListener(syncGameplay);

  function saveCloud(flush) {
    clearTimeout(cloudSaveTimer);
    cloudSaveTimer = 0;
    if (player && cloud) {
      call(() => player.setData(cloud, flush));
    }
  }

  function scheduleCloudSave() {
    clearTimeout(cloudSaveTimer);
    cloudSaveTimer = setTimeout(() => saveCloud(false), cloudSaveDelayMs);
  }

  // Уход со страницы — отложенное сохранение уходит сразу
  document.addEventListener("visibilitychange", () => {
    if (document.hidden && cloudSaveTimer) {
      saveCloud(true);
    }
  });

  async function loadCloud() {
    try {
      const loadedPlayer = await ysdk.getPlayer({ scopes: false });
      const data = await loadedPlayer.getData();
      // Игра уже читала сохранения: опоздавшее облако не принимаем
      // и не перезаписываем — до следующего запуска
      if (storageTouched) {
        return;
      }
      player = loadedPlayer;
      cloud = data && typeof data === "object" ? data : {};
    } catch (_) {
      player = null;
      cloud = null;
    }
  }

  platform.init = async function () {
    if (!window.YaGames) {
      return;
    }
    try {
      ysdk = await window.YaGames.init();
    } catch (_) {
      return;
    }
    // Язык игры — язык интерфейса Яндекс Игр (ISO 639-1); определяется
    // до запуска игры, boot.js ждёт init
    call(() => platform.setLanguage(ysdk.environment.i18n.lang));
    call(() => ysdk.on("game_api_pause", () => platform.suspend("portal")));
    call(() => ysdk.on("game_api_resume", () => platform.resume("portal")));
    platform.rewardAvailable = true;
    await loadCloud();
    // Игра могла стартовать раньше, чем SDK ответил
    sendReady();
    syncGameplay();
  };

  platform.ready = function () {
    readyRequested = true;
    sendReady();
  };

  platform.gameplay = function (active) {
    gameplayRequested = Boolean(active);
    syncGameplay();
  };

  platform.intermission = function () {
    if (!ysdk) {
      return;
    }
    platform.suspend("ad");
    const done = () => platform.resume("ad");
    try {
      ysdk.adv.showFullscreenAdv({
        callbacks: { onClose: done, onError: done, onOffline: done },
      });
    } catch (_) {
      done();
    }
  };

  platform.requestReward = function () {
    if (!ysdk) {
      rewardStatus = "denied";
      return;
    }
    rewardStatus = "pending";
    let rewarded = false;
    const done = () => {
      if (rewardStatus === "pending") {
        rewardStatus = rewarded ? "granted" : "denied";
      }
      platform.resume("ad");
    };
    platform.suspend("ad");
    try {
      ysdk.adv.showRewardedVideo({
        callbacks: {
          onRewarded: () => { rewarded = true; },
          onClose: done,
          onError: done,
        },
      });
    } catch (_) {
      done();
    }
  };

  // Итоговый исход отдаётся один раз
  platform.rewardStatus = function () {
    const status = rewardStatus;
    if (status === "granted" || status === "denied") {
      rewardStatus = "";
    }
    return status;
  };

  // Облако главнее: локальные сохранения переносятся в него
  // при первом чтении и остаются запасной копией
  platform.storage = {
    getItem(key) {
      storageTouched = true;
      if (cloud && Object.prototype.hasOwnProperty.call(cloud, key)) {
        const value = String(cloud[key]);
        baseStorage.setItem(key, value);
        return value;
      }
      const value = baseStorage.getItem(key);
      if (cloud && value !== null) {
        cloud[key] = value;
        scheduleCloudSave();
      }
      return value;
    },
    setItem(key, value) {
      storageTouched = true;
      const saved = baseStorage.setItem(key, value);
      if (!cloud) {
        return saved;
      }
      cloud[key] = value;
      scheduleCloudSave();
      return true;
    },
  };
})();
