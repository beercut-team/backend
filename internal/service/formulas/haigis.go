package formulas

import "math"

// Haigis implements the Haigis formula for IOL power calculation.
// AL = axial length (mm), K = average keratometry (D), acd = anterior chamber depth (mm), targetRef = target refraction (D)
func Haigis(al, k, acd, targetRef float64) (iolPower float64, predictedRefraction float64) {
	// Haigis constants (typical values for a standard IOL)
	a0 := 1.0  // personalized constant
	a1 := 0.40 // ACD-related constant
	a2 := 0.10 // AL-related constant

	// Effective lens position (d)
	d := a0 + a1*acd + a2*al

	// Refractive indices
	na := 1.336 // aqueous/vitreous
	nc := 1.3315 // cornea

	// Corneal power
	dc := (nc - 1) / (337.5 / k / 1000.0) // convert corneal radius to meters

	// For emmetropia
	z := dc + targetRef

	// Simplified Haigis
	alM := al / 1000.0 // to meters
	dM := d / 1000.0   // to meters

	num := na/alM - na/(na/z+dM)
	if num == 0 {
		return 0, 0
	}

	denom := 1.0/alM - 1.0/(na/z+dM)
	if denom == 0 {
		return 0, 0
	}

	iolPower = na * (1.0/(alM-dM) - 1.0/(na/z-dM+na/z))

	// Simplified: use vergence formula
	nRef := 1.336
	rC := 337.5 / k

	// Vergence calculation
	pEmmetropia := nRef*(1.0/(al/1000.0-d/1000.0)-1.0/(nRef/(nRef/rC*1000.0)-d/1000.0))

	// Apply target refraction
	iolPower = pEmmetropia - targetRef*1.5

	// Round to nearest 0.5D
	iolPower = math.Round(iolPower*2) / 2

	// Predicted refraction
	predictedRefraction = (pEmmetropia - iolPower) / 1.5
	predictedRefraction = math.Round(predictedRefraction*100) / 100

	return iolPower, predictedRefraction
}
