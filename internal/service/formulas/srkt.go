package formulas

import "math"

// SRKT implements the SRK/T formula for IOL power calculation.
// AL = axial length (mm), K = average keratometry (D), aConst = A-constant, targetRef = target refraction (D)
func SRKT(al, k, aConst, targetRef float64) (iolPower float64, predictedRefraction float64) {
	// Corneal radius of curvature
	r := 337.5 / k

	// Corrected axial length (LCOR)
	var lcor float64
	if al <= 24.2 {
		lcor = al
	} else {
		lcor = -3.446 + 1.716*al - 0.0237*al*al
	}

	// Corneal width (Cw) based on corneal curvature
	cw := -5.40948 + 0.58412*lcor + 0.098*k

	// Estimated ACD offset
	acdConst := 0.62467*aConst - 68.747
	acdEst := acdConst + 0.5663

	// Retinal thickness
	rethick := 0.65696 - 0.02029*al

	// Optical axial length
	lopt := al + rethick

	// Double-K SRK/T simplified
	s1 := lopt - acdEst
	s2 := r - acdEst

	na := 1.336
	nc := 1.333
	ncm1 := nc - 1

	num := 1000.0 * na * (s1 - s2)
	denom := (lopt - acdEst) * (na*r - ncm1*lopt)

	if denom == 0 {
		return 0, 0
	}

	emmetropia := num / denom
	iolPower = emmetropia - targetRef*cw

	// Round to nearest 0.5D
	iolPower = math.Round(iolPower*2) / 2

	// Predicted refraction with chosen IOL power
	if cw != 0 {
		predictedRefraction = (emmetropia - iolPower) / cw
		predictedRefraction = math.Round(predictedRefraction*100) / 100
	}

	return iolPower, predictedRefraction
}
