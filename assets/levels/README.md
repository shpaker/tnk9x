# Levels and campaign

This folder holds the maps (`N.bcmap`) and the campaign (`main.bccamp`).
Both are plain text files. The game reads them at startup. If any
level in the campaign has an error, the game refuses to start and the
message names the file and the line.

## Map file: `N.bcmap`

`N` is the level number (`1.bcmap`, `2.bcmap`, …). A map file has
three sections:

```
# Comments start with # (except inside [map], where # is a brick)

[level]
name = FIRST BLOOD
max_active = 4
time_3star = 130

[waves]
# tanks        delay  start
BBBb            150   now
BFBBb           140   left<=1
PPAaA           130   clear

[map]
..........................
..##..##..##..##..##..##..
...
```

Empty lines are ignored everywhere.

### `[level]`: level settings

All keys are optional. Missing keys fall back to `game.default_level`
in `config.yml`.

| Key          | Meaning                                                                  |
|--------------|--------------------------------------------------------------------------|
| `name`       | Title shown on the stage select screen. Keep it within 15 characters. Default: `STAGE NN`. |
| `max_active` | How many enemies can be on the field at once, `1..10`. In two-player mode the game adds 2. |
| `time_3star` | Time limit in seconds for the third star (see "Stars" below).            |

### `[waves]`: enemy waves

Each line is one wave: `tanks [delay] [start]`.

**tanks** lists the enemies in the order they appear:

| Letter | Tank                                |
|--------|-------------------------------------|
| `B`    | Basic: slow, weak shots              |
| `F`    | Fast: fast hull                      |
| `P`    | Power: fast shots                    |
| `A`    | Armor: slow, takes 4 hits            |

A **lowercase** letter (`b`, `f`, `p`, `a`) is a bonus carrier. It
flashes, and destroying it drops a power-up. If no wave contains
a lowercase letter, carriers follow the classic numbering: enemies
number 4, 9, 15 and 22.

The level's enemy total is the sum of the tanks in all waves. The
sidebar icons show how many are still waiting to spawn.

**delay** (optional) is the pause between spawns inside this wave, in
ticks (60 ticks = 1 second). Default: `game.enemy_respawn_delay_ticks`.
The first tank of a wave waits the same pause after the previous
spawn.

**start** (optional, default `now`) says when the wave may begin. The
condition is checked only after every tank of the previous wave has
spawned:

| Value      | The wave starts when…                                           |
|------------|-----------------------------------------------------------------|
| `now`      | immediately (the waves merge into one stream)                    |
| `left<=N`  | at most `N` enemies from earlier waves are still alive           |
| `clear`    | every enemy from earlier waves is destroyed (a "boss" wave)     |

The first wave ignores `start`. When the level begins, the first
tanks of the first wave (up to 3, and no more than `max_active`)
appear at once, one per spawn point.

### `[map]`: the field

26 rows of 26 characters. One character is one 8×8 tile, and a tank
is 2×2 tiles.

| Char | Tile                                      |
|------|-------------------------------------------|
| `.`  | empty                                     |
| `#`  | brick (destructible)                      |
| `@`  | steel (only a fully upgraded tank breaks it) |
| `~`  | water (blocks tanks without a boat, not bullets) |
| `%`  | forest (hides tanks)                      |
| `-`  | ice (tanks slide)                         |

Keep these areas empty. They are fixed in `config.yml`, not in the map:

- enemy spawn points: the top-left, top-center and top-right 2×2
  cells (`game.enemy_spawners`);
- player spawn points on the bottom row (`game.players_1_spawn_at`,
  `game.players_2_spawn_at`);
- the HQ at the bottom center (`game.hq_position`), usually wrapped in
  bricks.

Every row must have the same length. Any other character is an error.

### Legacy maps

A file with no sections, just the 26×26 grid, still works. The whole
file is read as `[map]`, and waves and settings come from
`game.default_level` in `config.yml`.

### Spawn point choice

When an enemy is due, the game picks one of the free spawn points. The
choice is random but weighted:

- a point covered by a tank is skipped; if every point is covered, the
  game waits;
- the point used last is picked 4× less often, the one before it
  about 1.7× less often;
- a point with a player within `game.spawn_player_safe_radius` tank
  cells is picked about 7× less often, so enemies rarely appear right
  next to the player.

## Stars

A won level gives 1 to 3 stars:

1. ★ the level is won;
2. ★★ no lives were lost;
3. ★★★ no lives were lost and the level took no longer than
   `time_3star`.

The best result per level is saved: in the OS config folder on
desktop, in `localStorage` in the browser. Losing a level never lowers
the saved stars.

## Campaign file: `main.bccamp`

The campaign groups levels into packs and sets how packs unlock.
`config.yml` → `game.campaign` names the file (`main` →
`main.bccamp`).

```
[campaign]
name = MAIN

[pack]
name = BOOT CAMP
levels = 1-5
order = sequential

[pack]
name = FIRST CONTACT
levels = 6-10
order = sequential
unlock_stars = 10
unlock_after = 5
```

`[campaign]` comes first and has one key, `name`. Every `[pack]` adds
the next pack, in order:

| Key            | Meaning                                                                 |
|----------------|-------------------------------------------------------------------------|
| `name`         | Pack title on the stage select screen.                                   |
| `levels`       | Level numbers: a range `1-5`, a list `1,2,7`, or both `1-3,8`. Required. |
| `order`        | `sequential` (default): the pack opens with only its first level, and winning a level opens the next one. `any`: every level opens with the pack. |
| `unlock_stars` | Total campaign stars needed to open the pack. Default `0`.               |
| `unlock_after` | Level that must be won first, from an earlier pack. Optional.            |

A pack opens when **both** `unlock_stars` and `unlock_after` are met.
The first pack usually has neither, so on a new save only level 1 is
open.

Checks at startup:

- every listed level has a `.bcmap` file that parses;
- no level appears in two packs;
- `unlock_after` points to a level in an earlier pack.

### Current layout

| Pack | Name          | Levels | Opens with                    |
|------|---------------|--------|-------------------------------|
| 1    | BOOT CAMP     | 1–5    | from the start                 |
| 2    | FIRST CONTACT | 6–10   | stage 5 won and 10 stars       |
| 3    | HARD LINE     | 11–15  | stage 10 won and 20 stars      |
| 4    | IRON STORM    | 16–20  | stage 15 won and 32 stars      |
| 5    | COLD FRONT    | 21–25  | stage 20 won and 45 stars      |
| 6    | DEEP STRIKE   | 26–30  | stage 25 won and 60 stars      |
| 7    | LAST STAND    | 31–35  | stage 30 won and 75 stars      |

The campaign has 105 stars in total. Each threshold needs roughly
two stars per level from the earlier packs, so finishing with one star
everywhere is not enough to get far.

## Adding a level

1. Create `N.bcmap` with the three sections.
2. Add `N` to a pack in `main.bccamp`.
3. Start the game. A parse error stops startup and names the line.
   The stage select screen shows the minimap, the enemy count per type
   and the 3-star time.
