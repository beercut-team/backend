package formulas

import "math"

// HofferQ implements the Hoffer Q formula for IOL power calculation.
// AL = axial length (mm), K = average keratometry (D), acd = anterior chamber depth (mm), targetRef = target refraction (D)
func HofferQ(al, k, acd, targetRef float64) (iolPower float64, predictedRefraction float64) {
	// Personalized ACD (pACD)
	pACD := 5.41 // Hoffer Q constant
	if acd > 0 {
		// Use measured ACD if available
		pACD = acd + 0.3
	}

	// Tangent factor for corneal curvature
	tanK := math.Tan(k * 0.01745329) // convert degrees to radians approximation

	// ACD estimate based on AL
	var acdOffset float64
	if al < 23.0 {
		acdOffset = 0.3 * (23.0 - al)
	} else {
		acdOffset = 0.3 * (al - 23.0)
	}

	acdEst := pACD + acdOffset + tanK*0.1

	// Refractive index
	n := 1.336

	// Corneal radius in mm
	rC := 337.5 / k

	// IOL power for emmetropia
	pEmmetropia := n/(al/1000.0-acdEst/1000.0) - n/(n/((n-1.0)/(rC/1000.0))-acdEst/1000.0)

	// Vergence correction factor
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
