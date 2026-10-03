// Мост Яндекс Игр поверх базового (web/common/bridge.js): SDK, язык
// интерфейса Яндекс Игр, разметка загрузки и геймплея, межуровневая
// реклама и реклама за награду, пауза площадки, облачные сохранения
// и покупки, просьба оценить игру. Межуровневая реклама — не чаще раза
// в 5 минут игрового времени, оценку просим не чаще раза за запуск.
// Вне Яндекса (SDK не загрузился) остаётся поведение базового моста.
// Методы не бросают исключений
(function () {
  "use strict";

  const platform = window.tnk9xPlatform;
  const baseStorage = platform.storage;

  // Задержка отправки облачных сохранений: серию записей шлём одной
  const cloudSaveDelayMs = 1000;

  // Межуровневая реклама показывается, только если с запуска или
  // с прошлой рекламы набрано столько игрового времени: карта длится
  // 1–3 минуты, реклама — примерно раз в несколько карт
  const intermissionIntervalMs = 5 * 60 * 1000;

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

  // Игровое время с прошлой рекламы: накопленное и начало текущего
  // отрезка размеченного геймплея
  let playedMs = 0;
  let gameplayStartedAt = 0;

  // Облачные сохранения: ключ — строка игры; null — облака нет
  let player = null;
  let cloud = null;
  let storageTouched = false;
  let cloudSaveTimer = 0;

  // Исход рекламы за награду в контракте моста
  let rewardStatus = "";

  // Оценку игры Яндекс разрешает просить раз за сессию
  let ratingAsked = false;

  // Покупки: каталог { id, price }, покупки игрока { id, token }
  // и исход последней покупки в контракте моста
  let payments = null;
  let catalog = [];
  let purchases = [];
  let purchaseStatus = "";

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
    if (active) {
      gameplayStartedAt = performance.now();
    } else {
      playedMs += performance.now() - gameplayStartedAt;
    }
    call(() => active
      ? ysdk.features.GameplayAPI.start()
      : ysdk.features.GameplayAPI.stop());
  }
  platform.addSuspendListener(syncGameplay);

  // Игровое время с прошлой рекламы: меню, пауза и приостановка
  // площадкой не считаются — разметка геймплея в них снята
  function playedTime() {
    const current = gameplayMarked ? performance.now() - gameplayStartedAt : 0;
    return playedMs + current;
  }

  // Любая показанная реклама начинает отсчёт заново
  function resetAdTimer() {
    playedMs = 0;
    gameplayStartedAt = performance.now();
  }

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

  function toPurchase(purchase) {
    return {
      id: String(purchase.productID),
      token: String(purchase.purchaseToken),
    };
  }

  // Каталог и покупки игрока — до запуска игры: игра зачисляет
  // покупки, не списанные в прошлых запусках
  async function loadPayments() {
    try {
      const loaded = await ysdk.getPayments({ signed: false });
      const products = await loaded.getCatalog();
      const owned = await loaded.getPurchases();
      catalog = Array.from(products, (product) => ({
        id: String(product.id),
        price: String(product.price),
      }));
      purchases = Array.from(owned, toPurchase);
      payments = loaded;
      platform.purchasesAvailable = catalog.length > 0;
    } catch (_) {
      payments = null;
      platform.purchasesAvailable = false;
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
    await Promise.all([loadCloud(), loadPayments()]);
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
    if (!ysdk || playedTime() < intermissionIntervalMs) {
      return;
    }
    platform.suspend("ad");
    const done = () => platform.resume("ad");
    // Отсчёт заново — только если реклама была показана
    const closed = (wasShown) => {
      if (wasShown !== false) {
        resetAdTimer();
      }
      done();
    };
    try {
      ysdk.adv.showFullscreenAdv({
        callbacks: { onClose: closed, onError: done, onOffline: done },
      });
    } catch (_) {
      done();
    }
  };

  // Просьба оценить игру. Не поверх рекламы: если межуровневая
  // реклама только что началась, попробуем на следующей победе
  platform.rating = function () {
    if (!ysdk || ratingAsked || platform.suspended) {
      return;
    }
    ratingAsked = true;
    call(async () => {
      const { value } = await ysdk.feedback.canReview();
      if (!value) {
        return;
      }
      platform.suspend("rating");
      try {
        await ysdk.feedback.requestReview();
      } finally {
        platform.resume("rating");
      }
    });
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
    // Показанная реклама за награду тоже начинает отсчёт заново
    const closed = () => {
      resetAdTimer();
      done();
    };
    platform.suspend("ad");
    try {
      ysdk.adv.showRewardedVideo({
        callbacks: {
          onRewarded: () => { rewarded = true; },
          onClose: closed,
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

  platform.catalog = function () {
    return catalog;
  };

  platform.purchases = function () {
    return purchases;
  };

  // Окно оплаты Яндекса поверх игры: игра стоит, пока оно открыто
  platform.requestPurchase = function (id) {
    if (!payments) {
      purchaseStatus = "denied";
      return;
    }
    purchaseStatus = "pending";
    platform.suspend("purchase");
    const done = (status) => {
      purchaseStatus = status;
      platform.resume("purchase");
    };
    try {
      payments.purchase({ id: String(id) }).then(
        (purchase) => {
          purchases = purchases.concat(toPurchase(purchase));
          done("granted");
        },
        () => done("denied"),
      );
    } catch (_) {
      done("denied");
    }
  };

  // Итоговый исход отдаётся один раз
  platform.purchaseStatus = function () {
    const status = purchaseStatus;
    if (status === "granted" || status === "denied") {
      purchaseStatus = "";
    }
    return status;
  };

  // Расходуемая покупка зачислена игрой: у Яндекса её больше нет
  platform.consumePurchase = function (token) {
    purchases = purchases.filter((purchase) => purchase.token !== token);
    if (payments) {
      call(() => payments.consumePurchase(String(token)));
    }
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
