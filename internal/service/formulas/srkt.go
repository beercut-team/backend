package formulas

import "math"

// SRKT implements the SRK/T formula for IOL power calculation.
// AL = axial length (mm), K = average keratometry (D), aConst = A-constant, targetRef = target refraction (D)
func SRKT(al, k, aConst, targetRef float64) (iolPower float64, predictedRefraction float64) {
	// Corneal radius of curvature in mm
	rC := 337.5 / k

	// Corrected axial length (LCOR)
	var lcor float64
	if al <= 24.2 {
		lcor = al
	} else {
		lcor = -3.446 + 1.716*al - 0.0237*al*al
	}

	// Corneal width/height factor
	cw := -5.40948 + 0.58412*lcor + 0.098*k

	// ACD constant calculation
	acdConst := 0.62467*aConst - 68.747

	// Retinal thickness
	rethick := 0.65696 - 0.02029*al

	// Optical axial length
	lopt := al + rethick

	// Effective lens position (ELP)
	elp := acdConst - 3.3357 + 0.1316*lopt + 0.0976*k

	// Refractive index
	n := 1.336

	// IOL power for emmetropia using vergence formula
	// P = n/(AL-ELP) - n/(n/K-ELP)
	pEmmetropia := n/(lopt/1000.0-elp/1000.0) - n/(n/((n-1.0)/(rC/1000.0))-elp/1000.0)

	// Apply target refraction correction
	iolPower = pEmmetropia - targetRef*cw

	// Round to nearest 0.5D
	iolPower = math.Round(iolPower*2) / 2

	// Predicted refraction with chosen IOL power
	if cw != 0 {
		predictedRefraction = (pEmmetropia - iolPower) / cw
		predictedRefraction = math.Round(predictedRefraction*100) / 100
	}

	return iolPower, predictedRefraction
}
