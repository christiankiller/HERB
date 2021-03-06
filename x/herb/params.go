package herb

import (
	"encoding/binary"

	"github.com/corestario/HERB/x/herb/types"
	"go.dedis.ch/kyber/v3"
	kyberenc "go.dedis.ch/kyber/v3/util/encoding"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//this file defines HERB parameters (such as threshold, participants ID's etc.) functions

// SetKeyHoldersNumber set the number of key holders (n for (t, n)-threshold cryptosystem)
func (k *Keeper) SetKeyHoldersNumber(ctx sdk.Context, n uint64) {
	store := ctx.KVStore(k.storeKey)
	nBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nBytes, n)
	store.Set([]byte(keyKeyHoldersNumber), nBytes)
}

// GetKeyHoldersNumber returns size of the current key holders group
func (k *Keeper) GetKeyHoldersNumber(ctx sdk.Context) (uint64, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(keyKeyHoldersNumber)) {
		return 0, sdk.ErrUnknownRequest("Store doesn't contain number of key holders")
	}
	nBytes := store.Get([]byte(keyKeyHoldersNumber))
	n := binary.LittleEndian.Uint64(nBytes)
	return n, nil
}

// SetVerificationKeys set verification keys corresponding to each address
func (k *Keeper) SetVerificationKeys(ctx sdk.Context, verificationKeys []types.VerificationKeyJSON) sdk.Error {
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(keyVerificationKeys)) {
		return sdk.ErrUnknownRequest("verification keys already exist")
	}

	verificationKeysBytes, err := k.cdc.MarshalJSON(verificationKeys)
	if err != nil {
		return sdk.ErrUnknownRequest("can't marshal list")
	}

	store.Set([]byte(keyVerificationKeys), verificationKeysBytes)
	return nil

}

// InitializeVerificationKeys runs at the start of the HERB protocol
// It's purpose is get verification keys from the KVStore and save them as a keeper field
// By default we can get gas limit problem with storing verification keys only in KVStore
func (k *Keeper) InitializeVerificationKeys(ctx sdk.Context) sdk.Error {
	verificationKeysJSON, err := k.GetVerificationKeys(ctx)
	if err != nil {
		return err
	}
	verificationKeys, err := types.VerificationKeyArrayDeserialize(verificationKeysJSON)
	if err != nil {
		return err
	}
	vk := make(map[string]types.VerificationKey)
	for _, key := range verificationKeys {
		vk[key.Sender.String()] = *key
	}
	k.verificationKeys = vk
	return nil
}

// GetVerificationKeys returns verification keys corresponding to each address
func (k *Keeper) GetVerificationKeys(ctx sdk.Context) ([]types.VerificationKeyJSON, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(keyVerificationKeys)) {
		return nil, sdk.ErrUnknownRequest("Verification keys are not defined")
	}
	verificationKeysBytes := store.Get([]byte(keyVerificationKeys))
	var verificationKeys []types.VerificationKeyJSON
	k.cdc.MustUnmarshalJSON(verificationKeysBytes, &verificationKeys)
	return verificationKeys, nil
}

// SetThreshold set threshold for decryption and ciphertext shares
func (k *Keeper) SetThreshold(ctx sdk.Context, thresholdCiphertexts uint64, thresholdDecrypt uint64) {
	store := ctx.KVStore(k.storeKey)
	thresholdCiphertextsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(thresholdCiphertextsBytes, thresholdCiphertexts)
	store.Set([]byte(keyThresholdCiphertexts), thresholdCiphertextsBytes)
	thresholdDecryptBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(thresholdDecryptBytes, thresholdDecrypt)
	store.Set([]byte(keyThresholdDecrypt), thresholdDecryptBytes)
}

// GetThresholdCiphertexts returns the total number of ciphertexts which required by HERB settings
func (k *Keeper) GetThresholdCiphertexts(ctx sdk.Context) (uint64, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(keyThresholdCiphertexts)) {
		return 0, sdk.ErrUnknownRequest("threshold for ciphertext shares is not defined")
	}

	tBytes := store.Get([]byte(keyThresholdCiphertexts))
	t := binary.LittleEndian.Uint64(tBytes)
	return t, nil
}

// GetThresholdDecryption returns threshold value for ElGamal cryptosystem
func (k *Keeper) GetThresholdDecryption(ctx sdk.Context) (uint64, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(keyThresholdDecrypt)) {
		return 0, sdk.ErrUnknownRequest("decryption threshold is not defined")
	}

	tBytes := store.Get([]byte(keyThresholdDecrypt))
	t := binary.LittleEndian.Uint64(tBytes)
	return t, nil
}

func (k *Keeper) SetCommonPublicKey(ctx sdk.Context, pubKeyHex string) {
	store := ctx.KVStore(k.storeKey)
	keyBytes := []byte(pubKeyHex)
	store.Set([]byte(keyCommonKey), keyBytes)
}

func (k *Keeper) GetCommonPublicKey(ctx sdk.Context) (kyber.Point, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	keyBytes := store.Get([]byte(keyCommonKey))
	key, err := kyberenc.StringHexToPoint(P256, string(keyBytes))
	if err != nil {
		return nil, sdk.ErrUnknownRequest("common key is not defined")
	}
	return key, nil
}
