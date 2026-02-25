package formulas

import "math"

// HofferQ implements the Hoffer Q formula for IOL power calculation.
// AL = axial length (mm), K = average keratometry (D), acd = anterior chamber depth (mm), targetRef = target refraction (D)
func HofferQ(al, k, acd, targetRef float64) (iolPower float64, predictedRefraction float64) {
	// Personalized ACD (pACD)
	pACD := acd
	if pACD == 0 {
		pACD = 3.336 // default
	}

	// Corrected axial length for Hoffer Q
	var m float64
	if al <= 23.0 {
		m = 1.0
	} else if al > 23.0 {
		m = -1.0
	}

	g := 28.4 - al/10.0

	// ACD estimate
	acdEst := pACD + 0.3*(al-23.5) + math.Tan(k*math.Pi/180.0)*0.1*math.Pow(23.5-al, 2)*m - 0.99166*g*0.1

	if acdEst < 2.5 {
		acdEst = 2.5
	}

	// Refractive index
	na := 1.336

	// Vergence-based IOL power
	rC := 337.5 / k // corneal radius of curvature (mm)

	// Effective optical length
	opticalAL := al

	// IOL power for emmetropia
	num1 := na / (opticalAL/1000.0 - acdEst/1000.0)
	num2 := na / (na/(na/rC*1000.0) - acdEst/1000.0)

	pEmmetropia := num1 - num2

	// Simplified Hoffer Q approximation
	pEmmetropia = 1336.0/(al-acdEst-0.05) - 1336.0/(1336.0/((1000.0/((1000.0/k)-rC+0.05))+0.001)-acdEst-0.05)

	// Apply target refraction correction
	iolPower = pEmmetropia - targetRef*1.5

	// Round to nearest 0.5D
	iolPower = math.Round(iolPower*2) / 2

	// Predicted refraction
	predictedRefraction = (pEmmetropia - iolPower) / 1.5
	predictedRefraction = math.Round(predictedRefraction*100) / 100

	return iolPower, predictedRefraction
}
