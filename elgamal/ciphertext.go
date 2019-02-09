package elgamal

import (
	"crypto/elliptic"
	"math/big"

	"github.com/dgamingfoundation/HERB/point"
)

//Ciphertext is usual ElGamal ciphertext C = (a, b)
//Here a, b - the elliptic curve's points
type Ciphertext struct {
	pointA point.Point
	pointB point.Point
}

//IdentityCiphertext creates ciphertext which is neutral with respect to plaintext group operation (after ciphertext aggregation operation)
func IdentityCiphertext(curve elliptic.Curve) Ciphertext {
	return Ciphertext{point.PointAtInfinity(curve), point.PointAtInfinity(curve)}
}

//Decrypt takes decrypt parts (shares) from all participants and decrypt the ciphertext C
func (ct Ciphertext) Decrypt(curve elliptic.Curve, shares []point.Point) point.Point {
	if len(shares) == 0 {
		return point.PointAtInfinity(curve)
	}

	//aggregating all parts
	decryptKey := shares[0]
	for i := 1; i < len(shares); i++ {
		decryptKey = decryptKey.Add(curve, shares[i])
	}

	M := ct.pointB.Add(curve, decryptKey.Neg(curve))
	return M
}

//IsValid return true if both part of the ciphertext C on the curve E.
func (ct Ciphertext) IsValid(curve elliptic.Curve) bool {
	statement1 := curve.IsOnCurve(ct.pointA.GetX(), ct.pointA.GetY())
	statement2 := curve.IsOnCurve(ct.pointB.GetX(), ct.pointB.GetY())
	return statement1 && statement2
}

//Equal compares two ciphertexts and returns true if ct = ct1
func (ct Ciphertext) Equal(ct1 Ciphertext) bool {
	return ct.pointA.Equal(ct1.pointA) && ct.pointB.Equal(ct1.pointB)
}

//AggregateCiphertext takes the set of ciphertextes parts:
//parts[0] = (A0, B0), ..., parts[n] = (An, Bn)
//and returns aggregated ciphertext C = (A1 + A2 + ... + An, B1 + B2 + ... + Bn)
func AggregateCiphertext(curve elliptic.Curve, parts []Ciphertext) Ciphertext {
	if len(parts) == 0 {
		return IdentityCiphertext(curve)
	}

	ct := Ciphertext{parts[0].pointA, parts[0].pointB}
	for i := 1; i < len(parts); i++ {
		ct.pointA = ct.pointA.Add(curve, parts[i].pointA)
		ct.pointB = ct.pointB.Add(curve, parts[i].pointB)
	}

	return ct
}

//Decrypt the ciphertext C with the key x
//Currently not in use
func decrypt(curve elliptic.Curve, ct Ciphertext, x *big.Int) point.Point {
	pointTemp := ct.pointA.ScalarMult(curve, x)
	pointTempY := pointTemp.GetY()
	pointTempY.Neg(pointTempY)

	//M = b - xA
	return ct.pointB.Add(curve, pointTemp)
}