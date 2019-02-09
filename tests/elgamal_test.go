package tests

import (
	"bytes"
	"crypto/elliptic"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"

	. "github.com/dgamingfoundation/HERB/elgamal"
	"github.com/dgamingfoundation/HERB/point"
)

func Test_ElGamal_Positive(t *testing.T) {
	testCases := []int{1, 2, 3, 5, 10, 50, 100, 300}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("validators set %d", tc), func(t *testing.T) {
			parties, curve := initElGamal(t, tc)
			elGamalPositive(t, parties, curve)
		})
	}
}

func Test_PointAtInfinity_Positive(t *testing.T) {
	curve := elliptic.P256()
	curveParams := curve.Params()

	genPoint, err := point.FromCoordinates(curve, curveParams.Gx, curveParams.Gy)
	if err != nil {
		t.Errorf("can't make genPoint: %s", err)
	}

	n1 := big.NewInt(1)
	n1.Sub(curveParams.N, big.NewInt(1))

	testCases := []point.Point{point.PointAtInfinity(curve), genPoint,
		genPoint.ScalarMult(curve, big.NewInt(13)), genPoint.ScalarMult(curve, n1)}

	pointInf := point.PointAtInfinity(curve)
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Scalar multiplication, point %d:", i), func(t *testing.T) {
			scalarMultPositive(t, curve, tc, pointInf)
		})
		t.Run(fmt.Sprintf("Addition, point %d:", i), func(t *testing.T) {
			addPositive(t, curve, tc, pointInf)
		})
		t.Run(fmt.Sprintf("Substraction (point at infinity), point %d:", i), func(t *testing.T) {
			subPositive(t, curve, tc, pointInf)
		})
		t.Run(fmt.Sprintf("Substraction (two equal points), point %d:", i), func(t *testing.T) {
			subTwoEqualPositive(t, curve, tc, pointInf)
		})
	}
}

func Test_IdentityCiphertext_Positive(t *testing.T) {
	curve := elliptic.P256()
	party, err := DKG(curve, 1)
	if err != nil {
		t.Errorf("can't init DKG with error %q", err)
	}

	genPoint, err := point.FromCoordinates(curve, curve.Params().Gx, curve.Params().Gy)
	if err != nil {
		t.Errorf("can't make genPoint: %s", err)
	}

	n1 := big.NewInt(1)
	n1.Sub(curve.Params().N, big.NewInt(1))

	messages := []point.Point{point.PointAtInfinity(curve), genPoint,
		genPoint.ScalarMult(curve, big.NewInt(13)), genPoint.ScalarMult(curve, n1)}

	testCases := make([]Ciphertext, len(messages))
	for i, m := range messages {
		testCases[i] = party[0].Encrypt(curve, m)
	}

	neutralCiphertext := IdentityCiphertext(curve)
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Ciphertext %d:", i), func(t *testing.T) {
			neutralCiphertextAggregate(t, curve, tc, neutralCiphertext, party[0])
		})
	}
}

func neutralCiphertextAggregate(t *testing.T, curve elliptic.Curve, ct Ciphertext, neutral Ciphertext, party Participant) {
	parts := []Ciphertext{ct, neutral}
	resultCT := AggregateCiphertext(curve, parts)

	originalDecryptShares := []point.Point{party.PartialDecrypt(curve, ct)}
	plaintext := ct.Decrypt(curve, originalDecryptShares)

	newDecryptShares := []point.Point{party.PartialDecrypt(curve, resultCT)}
	resultPlaintext := resultCT.Decrypt(curve, newDecryptShares)

	deepEqual(t, resultPlaintext, plaintext)
}

func scalarMultPositive(t *testing.T, curve elliptic.Curve, p point.Point, pointInf point.Point) {
	curveParams := curve.Params()
	multResult := p.ScalarMult(curve, curveParams.N)
	deepEqual(t, pointInf, multResult)
}

func addPositive(t *testing.T, curve elliptic.Curve, p point.Point, pointInf point.Point) {
	addResult := p.Add(curve, pointInf)
	deepEqual(t, p, addResult)
}

func subPositive(t *testing.T, curve elliptic.Curve, p point.Point, pointInf point.Point) {
	subResult := p.Sub(curve, pointInf)
	deepEqual(t, p, subResult)
}

func subTwoEqualPositive(t *testing.T, curve elliptic.Curve, p point.Point, pointInf point.Point) {
	fmt.Println(p.GetX(), p.GetY())
	subResult := p.Sub(curve, p)
	fmt.Println(p.GetX(), p.GetY(), subResult.GetX(), subResult.GetY())
	deepEqual(t, pointInf, subResult)
}

func elGamalPositive(t *testing.T, parties []Participant, curve elliptic.Curve) {
	n := len(parties)

	//Any system user generates some message, encrypts and publishes it
	//We use our validators set (parties) just for example
	publishedCiphertextes := make([]Ciphertext, n)

	newMessages := make([]point.Point, n)
	publishChan := publishMessages(parties, curve)
	for publishedMessage := range publishChan {
		i := publishedMessage.id

		newMessages[i] = publishedMessage.msg
		publishedCiphertextes[i] = publishedMessage.published
	}

	for i := range publishedCiphertextes {
		if !publishedCiphertextes[i].IsValid(curve) {
			t.Errorf("Ciphertext is not valid: %v\nOriginal message: %v", publishedCiphertextes[i], newMessages[i])
		}
	}

	//aggregate all ciphertextes
	commonCiphertext := AggregateCiphertext(curve, publishedCiphertextes)

	if !commonCiphertext.IsValid(curve) {
		t.Errorf("Common ciphertext is not valid: %v\nOriginal messages: %v", commonCiphertext, newMessages)
	}

	//decrypt the random
	decryptParts := make([]point.Point, n)
	decrypted := decryptMessages(parties, curve, commonCiphertext)
	for msg := range decrypted {
		i := msg.id
		decryptParts[i] = msg.msg
	}

	decryptedMessage := commonCiphertext.Decrypt(curve, decryptParts)

	expectedMessage, err := point.Recover(curve, newMessages)
	if err != nil {
		t.Errorf("can't recover the point with error: %q", err)
	}

	deepEqual(t, decryptedMessage, expectedMessage)
}

type errorf interface {
	Errorf(format string, args ...interface{})
}

func deepEqual(t errorf, obtained, expected interface{}) {
	if !cmp.Equal(obtained, expected) {
		t.Errorf("... %s", diff(obtained, expected))
	}
}

func diff(obtained, expected interface{}) string {
	var failMessage bytes.Buffer
	diffs := pretty.Diff(obtained, expected)

	if len(diffs) > 0 {
		failMessage.WriteString("Obtained:\t\tExpected:")
		for _, singleDiff := range diffs {
			failMessage.WriteString(fmt.Sprintf("\n%v", singleDiff))
		}
	}

	return failMessage.String()
}

func initElGamal(t errorf, n int) ([]Participant, elliptic.Curve) {
	// creating elliptic curve
	curve := elliptic.P256()

	//generating key
	parties, err := DKG(curve, n)
	if err != nil {
		t.Errorf("can't init DKG with error %q", err)
	}

	return parties, curve
}

type publishedMessage struct {
	id        int
	msg       point.Point
	published Ciphertext
}

func publishMessages(parties []Participant, curve elliptic.Curve) chan publishedMessage {
	publish := make(chan publishedMessage, len(parties))

	wg := sync.WaitGroup{}

	go func() {
		wg.Add(len(parties))

		for i := range parties {
			go func(id int) {
				message := point.New(curve)
				encryptedMessage := parties[id].Encrypt(curve, *message)

				publish <- publishedMessage{id, *message, encryptedMessage}
				wg.Done()
			}(i)
		}

		wg.Wait()
		close(publish)
	}()

	return publish
}

type decryptedMessage struct {
	id  int
	msg point.Point
}

func decryptMessages(parties []Participant, curve elliptic.Curve, commonCiphertext Ciphertext) chan decryptedMessage {
	decrypted := make(chan decryptedMessage, len(parties))

	wg := sync.WaitGroup{}

	go func() {
		wg.Add(len(parties))

		for i := range parties {
			go func(id int) {
				decryptedMsg := parties[id].PartialDecrypt(curve, commonCiphertext)

				decrypted <- decryptedMessage{id, decryptedMsg}
				wg.Done()
			}(i)
		}

		wg.Wait()
		close(decrypted)
	}()

	return decrypted
}