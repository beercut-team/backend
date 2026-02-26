package formulas

import "math"

// Haigis implements the Haigis formula for IOL power calculation.
// AL = axial length (mm), K = average keratometry (D), acd = anterior chamber depth (mm), targetRef = target refraction (D)
func Haigis(al, k, acd, targetRef float64) (iolPower float64, predictedRefraction float64) {
	// Haigis constants (typical values for a standard IOL)
	a0 := -1.0  // personalized constant
	a1 := 0.40  // ACD-related constant
	a2 := 0.10  // AL-related constant

	// Effective lens position (d) in mm
	d := a0 + a1*acd + a2*al

	// Refractive index
	n := 1.336

	// Corneal radius in mm
	rC := 337.5 / k

	// Corneal power in diopters
	dC := (n - 1.0) / (rC / 1000.0)

	// IOL power for emmetropia
	pEmmetropia := n/(al/1000.0-d/1000.0) - n/(n/dC-d/1000.0)

	// Vergence correction factor (simplified)
	vf := 1.5

	// Apply target refraction
	iolPower = pEmmetropia - targetRef*vf

	// Round to nearest 0.5D
	iolPower = math.Round(iolPower*2) / 2

	// Predicted refraction
	if vf != 0 {
		predictedRefraction = (pEmmetropia - iolPower) / vf
		predictedRefraction = math.Round(predictedRefraction*100) / 100
	}

	return iolPower, predictedRefraction
}
