package main

type UsableParts struct {
	Cabins  []*Cabin
	Wings   []*Wing
	Wheels  []*Wheel
	Engines []*Engine
}

func Repair(oldNest *Nest, newNest *Nest) {
	uParts := UsableParts{}
	for _, ac := range oldNest.Aircrafts {
		uParts.Cabins = append(uParts.Cabins, ac.Cabin)
		for _, w := range ac.Wings {
			uParts.Wings = append(uParts.Wings, w)
		}
		for _, wh := range ac.Wheels {
			uParts.Wheels = append(uParts.Wheels, wh)
		}
	}
	// Cabins
	for _, ac := range newNest.Aircrafts {
		if ac.Cabin == nil {
			if len(uParts.Cabins) > 0 {
				cabinPart := uParts.Cabins[0]
				ac.Cabin = cabinPart
				uParts.Cabins = uParts.Cabins[1:]
				//
				for _, oldAc := range oldNest.Aircrafts {
					if oldAc.Cabin == cabinPart {
						oldAc.Cabin = nil
						break
					}
				}
			}
		}
		// Wings
		// Important Go slice iteration rule:
		//
		// When iterating over a slice using `for _, v := range slice`,
		// the variable `v` is **a copy** of the slice element, not a reference.
		// So:
		//
		//   v == slice[i]    → comparison works fine ✔
		//   v = newValue     → modifies only the local copy ❌ (NO effect on slice)
		//
		// If you need to **modify** the real slice element,
		// always use the index form:
		//
		//   for i, v := range slice {
		//       slice[i] = newValue  // ✔ this actually mutates the slice
		//   }
		//
		// TL;DR:
		// - Use `v` for reading & comparison
		// - Use `slice[i]` for writing & mutation

		for i, w := range ac.Wings {
			if w == nil && len(uParts.Wings) > 0 {
				wingPart := uParts.Wings[0]
				ac.Wings[i] = wingPart
				uParts.Wings = uParts.Wings[1:]
				//
				for _, oldAc := range oldNest.Aircrafts {
					for oldIdx, oldWings := range oldAc.Wings {
						if oldWings == wingPart {
							oldAc.Wings[oldIdx] = nil
							break
						}
					}
				}
			}
		}
		// Wheels
		for i, wh := range ac.Wheels {
			if wh == nil && len(uParts.Wheels) > 0 {
				wheelPart := uParts.Wheels[0]
				ac.Wheels[i] = wheelPart
				uParts.Wheels = uParts.Wheels[1:]
				//
				for _, oldAc := range oldNest.Aircrafts {
					for oldIdx, oldWheel := range oldAc.Wheels {
						if oldWheel == wheelPart {
							oldAc.Wheels[oldIdx] = nil
							break
						}
					}
				}
			}
		}

	}
	// Update Engine
	uParts.Engines = nil
	for _, oldAc := range oldNest.Aircrafts {
		for _, w := range oldAc.Wings {
			if w != nil && w.Engine != nil {
				uParts.Engines = append(uParts.Engines, w.Engine)
			}
		}
	}
	// Engine
	for _, ac := range newNest.Aircrafts {
		for i, w := range ac.Wings {
			if w != nil && w.Engine == nil && len(uParts.Engines) > 0 {
				enginePart := uParts.Engines[0]
				ac.Wings[i].Engine = enginePart
				uParts.Engines = uParts.Engines[1:]
				//
				for _, oldAc := range oldNest.Aircrafts {
					for oldIdx, oldWheel := range oldAc.Wings {
						if oldWheel != nil && oldWheel.Engine == enginePart {
							oldAc.Wings[oldIdx].Engine = nil
							break
						}
					}
				}
			}
		}
	}
}
