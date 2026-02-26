package formulas

import (
	"math"
	"testing"
)

func TestSRKT(t *testing.T) {
	// Test case: typical eye parameters
	al := 23.5      // axial length in mm
	k := 44.0       // average keratometry in D
	aConst := 118.4 // A-constant
	targetRef := 0.0 // target emmetropia

	iolPower, predictedRef := SRKT(al, k, aConst, targetRef)

	// IOL power should be reasonable (typically 18-24D for normal eyes)
	if iolPower < 10.0 || iolPower > 30.0 {
		t.Errorf("SRKT IOL power out of reasonable range: %.2f D", iolPower)
	}

	// Predicted refraction should be close to target
	if math.Abs(predictedRef-targetRef) > 1.0 {
		t.Errorf("SRKT predicted refraction %.2f D too far from target %.2f D", predictedRef, targetRef)
	}

	// IOL power should be rounded to 0.5D
	if math.Mod(iolPower*2, 1.0) != 0 {
		t.Errorf("SRKT IOL power not rounded to 0.5D: %.2f", iolPower)
	}

	t.Logf("SRKT: AL=%.2f, K=%.2f, IOL=%.2f D, Predicted Ref=%.2f D", al, k, iolPower, predictedRef)
}

func TestHaigis(t *testing.T) {
	al := 23.5
	k := 44.0
	acd := 3.2 // anterior chamber depth
	targetRef := 0.0

	iolPower, predictedRef := Haigis(al, k, acd, targetRef)

	if iolPower < 10.0 || iolPower > 30.0 {
		t.Errorf("Haigis IOL power out of reasonable range: %.2f D", iolPower)
	}

	if math.Abs(predictedRef-targetRef) > 1.0 {
		t.Errorf("Haigis predicted refraction %.2f D too far from target %.2f D", predictedRef, targetRef)
	}

	if math.Mod(iolPower*2, 1.0) != 0 {
		t.Errorf("Haigis IOL power not rounded to 0.5D: %.2f", iolPower)
	}

	t.Logf("Haigis: AL=%.2f, K=%.2f, ACD=%.2f, IOL=%.2f D, Predicted Ref=%.2f D", al, k, acd, iolPower, predictedRef)
}

func TestHofferQ(t *testing.T) {
	al := 23.5
	k := 44.0
	acd := 3.2
	targetRef := 0.0

	iolPower, predictedRef := HofferQ(al, k, acd, targetRef)

	if iolPower < 10.0 || iolPower > 30.0 {
		t.Errorf("HofferQ IOL power out of reasonable range: %.2f D", iolPower)
	}

	if math.Abs(predictedRef-targetRef) > 1.0 {
		t.Errorf("HofferQ predicted refraction %.2f D too far from target %.2f D", predictedRef, targetRef)
	}

	if math.Mod(iolPower*2, 1.0) != 0 {
		t.Errorf("HofferQ IOL power not rounded to 0.5D: %.2f", iolPower)
	}

	t.Logf("HofferQ: AL=%.2f, K=%.2f, ACD=%.2f, IOL=%.2f D, Predicted Ref=%.2f D", al, k, acd, iolPower, predictedRef)
}

func TestShortEye(t *testing.T) {
	// Short eye (high hyperopia)
	al := 21.0
	k := 45.0
	aConst := 118.4
	targetRef := 0.0

	iolPower, _ := SRKT(al, k, aConst, targetRef)

	// Short eyes need higher IOL power
	if iolPower < 25.0 {
		t.Errorf("Short eye should have high IOL power, got %.2f D", iolPower)
	}

	t.Logf("Short eye: AL=%.2f, IOL=%.2f D", al, iolPower)
}

func TestLongEye(t *testing.T) {
	// Long eye (high myopia)
	al := 28.0
	k := 43.0
	aConst := 118.4
	targetRef := 0.0

	iolPower, _ := SRKT(al, k, aConst, targetRef)

	// Long eyes need lower IOL power
	if iolPower > 15.0 {
		t.Errorf("Long eye should have low IOL power, got %.2f D", iolPower)
	}

	t.Logf("Long eye: AL=%.2f, IOL=%.2f D", al, iolPower)
}
