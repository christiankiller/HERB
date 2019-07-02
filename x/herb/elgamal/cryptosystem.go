package elgamal

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/proof/dleq"
	"go.dedis.ch/kyber/v3/util/random"
)

// RandomCiphertext creates an elgamal ciphertext with a random plaintext
func RandomCiphertext(group proof.Suite, commonKey kyber.Point) (ct Ciphertext, DLKproof []byte, RKproof []byte, err error) {
	y := group.Scalar().Pick(random.New())
	M := group.Point().Mul(y, nil)
	r := group.Scalar().Pick(random.New())
	S := group.Point().Mul(r, commonKey)
	A := group.Point().Mul(r, nil)
	B := S.Add(group.Point().Mul(r, commonKey), M)
	ct = Ciphertext{A, B}
	DLKproof, err = DLK(group, group.Point().Base(), r, ct.PointA)
	if err != nil {
		return
	}
	RKproof, err = RK(group, group.Point().Base(), y, commonKey, r, ct.PointB)
	return
}

// create decryption shares and proof
func CreateDecShare(group proof.Suite, C Ciphertext, partKey kyber.Scalar) (decShare kyber.Point, DLEproof *dleq.Proof, err error) {
	decShare = group.Point().Mul(partKey, C.PointA)
	DLEproof, _, _, err = DLE(group, group.Point().Base(), C.PointA, partKey)
	if err != nil {
		return nil, nil, err
	}
	return
}
