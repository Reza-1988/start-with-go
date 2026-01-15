# Airplane Hangar (airplane-hanger)

A small Go kata about **repairing aircrafts by transplanting parts** from an “old” hangar into a “new” hangar.

You are given two `Nest`s:

- `oldNest`: a donor hangar containing aircrafts with usable parts
- `newNest`: a hangar containing aircrafts that may have missing/broken parts (`nil`)

Your job is to implement (or understand) `Repair(oldNest, newNest)` so the **new hangar becomes as repaired as possible**, using **only parts taken from the old hangar**.

---

## Data model

```go
type Nest struct {
    ID        int
    Aircrafts []*Aircraft
}

type Aircraft struct {
    ID     int
    Wings  []*Wing   // (in the tests: 2)
    Cabin  *Cabin    // (single)
    Wheels []*Wheel  // (in the tests: 3)
}

type Wing struct {
    ID     int
    Engine *Engine
}

type Cabin struct { ID int }
type Wheel struct { ID int }
type Engine struct { ID int }
````

A `nil` pointer means “missing / broken”.

---

## The rules of repair (the spec)

### 1) Only move existing parts (no creating new ones)

You **must not manufacture** new cabins/wings/wheels/engines.
Repairs happen by **moving pointers** from `oldNest` into `newNest`.

### 2) A moved part must be removed from `oldNest`

Whenever a part is transplanted into `newNest`, it must be **removed from its original location** in `oldNest` by setting that slot to `nil`.

This prevents the same physical part from being used twice.

### 3) Fill order matters (deterministic behavior)

Repairs are applied by scanning in this order:

1. Aircrafts in `newNest.Aircrafts` (from index `0` to end)
2. Within each aircraft, the part slots (slice order for wings/wheels)

And each time you need a part, you take the **first available** donor part from `oldNest` (like a queue / FIFO).

### 4) Repair priority matters (this is the tricky part)

Repair is performed in this priority order:

1. **Cabins**
2. **Wings**
3. **Wheels**
4. **Engines** (last)

Why engines last?

Because engines live *inside wings*. If you replace a missing wing, you are moving the wing object (and its engine, if it has one).
After all wing swaps are done, you only have engines available from whatever wings remain in the old hangar.

This is exactly why a wing’s engine problem might remain unsolved if wing swaps consume all donor wings first.

### 5) If there aren’t enough donor parts, leave `nil`

If `oldNest` runs out of cabins/wings/wheels/engines, the remaining missing parts in `newNest` stay `nil`.

---

## Example (mirrors the provided test)

* `oldNest` has **1 aircraft** fully equipped (1 cabin, 2 wings with engines, 3 wheels)
* `newNest` has **2 aircrafts**, but some parts are set to `nil` (missing cabin, missing wings, missing wheels, missing engine)

After `Repair(oldNest, newNest)`:

* The missing cabin is taken from `oldNest` and installed in `newNest`
* The missing wings are filled first (consuming donor wings)
* The missing wheels are filled until wheels run out
* The missing engine may remain missing if no engines are left after wing transfers

---

## How to run

### Run tests

```bash
go test ./...
```

This kata uses `testify/assert`, so the first test run may download dependencies automatically.

---

## Common Go pitfall (important)

When iterating over slices with:

```go
for _, v := range slice { ... }
```

`v` is a **copy** of the slice element. Assigning to `v` will not modify the slice.

If you need to mutate the slice (e.g., replace `nil` slots), iterate with indexes:

```go
for i, v := range slice {
    if v == nil {
        slice[i] = replacement
    }
}
```

---

## What “done” looks like

Your solution is correct when:

* `newNest` is repaired as much as possible
* moved parts are removed from `oldNest`
* behavior is deterministic and matches the tests (especially the **engine-after-wings** rule)

Happy repairing 

